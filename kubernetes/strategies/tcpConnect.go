package strategies

import (
	"Multiplexing_/kubernetes/resource"
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

/*
  - TCPConnectStrategy
  - 策略：目标追踪规则，目标是让容器的tcp连接数保持稳定
    根据容器的tcp连接数来确定是否扩容的策略
  - 方案1:
    针对激增的流量（tcp连接）可以指定三种判定策略，然后每次三种判定策略都会对当前的状态进行判定和投票
    如果投票认为当前激增的流量需要进行扩容的话 就扩容 否则不扩容
*/

const (
	strategiesNum              = 3                // 判断策略总数
	strategiesRecallNum        = 2                // 回收策略总数
	separateTime               = 3                // 设置监控的间隔时间段
	TimeSet                    = 5                // 每10秒统计一次
	averageGrowthRateThreshold = 2                // 流量平均增速 每次流量的增长量为2倍
	maxFloatRate               = 25               // 峰值流量
	loadFactor                 = float32(3 / 4)   // 负载因子
	maxScalingDeployments      = 3                // 最大deployment伸缩量	伸缩量*replicate < cpu核心数
	forbiddenExtendResource    = 5                // 禁止重复伸缩资源的时间
	TimeToRecallResource       = 2 * separateTime // 判断是否需要收缩pod（资源）的时间间隔
)

var (
	funcSlopesThreeMinuteVolume     int            // 窗口容量
	funcSingleTcpVolume             int            // 每次的tcp连接数量
	funcSlopesThreeMinute           []float32      // 三分钟内容器tcp连接的平均增长速率
	funcEachTcpThreeMinute          []int          // 三分钟内容器的tcp连接数量
	countSlopTimes                  bool           // 判断什么时候需要计算斜率
	wgExtend                        sync.WaitGroup // wg
	wgRecall                        sync.WaitGroup // wg
	chExtend                        chan bool      // vote投票者
	chRecall                        chan bool      // vote投票者
	hasExtendDeploymentInFiveMinute bool           // 5分钟内是否扩容过 默认没有
	scalingDeployments              []string       // 伸缩的deployment
	scalingDeploymentsLock          sync.Mutex     // 互斥锁
	hashExtendDeploymentLock        sync.RWMutex   // hasExtendDeploymentInFiveMinute状态修改的读写锁
)

// 初始化元素
func init() {
	funcSlopesThreeMinuteVolume = int(math.Ceil(time.Minute.Seconds()/float64(TimeSet))-1) * separateTime
	funcSingleTcpVolume = int(math.Ceil(time.Minute.Seconds()/float64(TimeSet))) * separateTime
	funcSlopesThreeMinute = make([]float32, 0, funcSlopesThreeMinuteVolume)
	funcEachTcpThreeMinute = make([]int, 0, funcSingleTcpVolume)
	countSlopTimes = false
	chExtend = make(chan bool, strategiesNum)
	chRecall = make(chan bool, strategiesRecallNum)
	hasExtendDeploymentInFiveMinute = false
}

type TCPConnectStrategy struct {
}

// ExtendDeploymentInFiveMinuteChange hashExtendDeploymentLock状态修改的唯一改变形式
// 修改状态变量 需要加锁 hasExtendDeploymentInFiveMinute变量修改是在go func的异步线程执行的所以需要加锁
// 保证主线程操作不受影响
func ExtendDeploymentInFiveMinuteChange() {
	hashExtendDeploymentLock.Lock()
	defer hashExtendDeploymentLock.Unlock()
	hasExtendDeploymentInFiveMinute = !hasExtendDeploymentInFiveMinute
}

// GetHasExtendDeploymentInFiveMinute 唯一获取hasExtendDeploymentInFiveMinute的方法
// 读取hasExtendDeploymentInFiveMinute 加入读写锁 异步写操作修改
func GetHasExtendDeploymentInFiveMinute() bool {
	hashExtendDeploymentLock.RLock()
	defer hashExtendDeploymentLock.RUnlock()
	return hasExtendDeploymentInFiveMinute
}

// RecallResource 回收资源
func (tcpStrategy *TCPConnectStrategy) RecallResource() string {
	scalingDeploymentsLock.Lock()
	defer scalingDeploymentsLock.Unlock()

	lastResource := scalingDeployments[len(scalingDeployments)-1]
	scalingDeployments = scalingDeployments[:len(scalingDeployments)-1]
	return lastResource
}

// CountingResourceExtendTime 保持资源伸缩后的状态
// 将状态进行切换 然后睡眠5分钟  保证5分钟内状态不会改变
func (tcpStrategy *TCPConnectStrategy) CountingResourceExtendTime() {
	log.Printf("改变资源伸缩后的状态，%d分钟 内不允许再次伸缩", forbiddenExtendResource)
	ExtendDeploymentInFiveMinuteChange()
	time.Sleep(time.Minute * time.Duration(forbiddenExtendResource))
	ExtendDeploymentInFiveMinuteChange()
	log.Printf("经过%d分钟时间后，允许再次对资源进行伸缩", forbiddenExtendResource)
}

// ExpandResource 单例模式 获取全局变量 并追加伸缩的资源Resource
func (tcpStrategy *TCPConnectStrategy) ExpandResource(resourceName string) {
	scalingDeploymentsLock.Lock()
	defer scalingDeploymentsLock.Unlock()

	if scalingDeployments == nil {
		scalingDeployments = make([]string, 0, maxScalingDeployments)
	}

	scalingDeployments = append(scalingDeployments, resourceName)
}

// GetExtendDeploymentNum 获取scalingDeployments 长度
func (tcpStrategy *TCPConnectStrategy) GetExtendDeploymentNum() int {
	scalingDeploymentsLock.Lock()
	defer scalingDeploymentsLock.Unlock()

	if scalingDeployments == nil {
		return 0
	}
	return len(scalingDeployments)
}

// CheckIfNeedDynamicExtend 检查是否需要扩容
func (tcpStrategy *TCPConnectStrategy) CheckIfNeedDynamicExtend(parameter interface{}) bool {
	tcpRes := parameter.(*resource.Tcp)
	tcpStrategy.recordTcpConnectionTreeMinute(tcpRes)

	wgExtend.Add(strategiesNum)
	go tcpStrategy.strategyFunctionCountMeanSlope()
	go tcpStrategy.strategyFunctionExceedFloatRate()
	go tcpStrategy.strategyFunctionMaxMinGap()
	wgExtend.Wait()

	fmt.Printf("vote res: ")
	voteY := 0
	voteN := 0
	for i := 0; i < strategiesNum; i++ {
		vote := <-chExtend
		if vote {
			voteY++
		} else {
			voteN++
		}
		fmt.Printf("%t ", vote)
	}
	fmt.Println()

	return voteY > voteN && tcpStrategy.checkIfCanExtend()
}

// CheckIfNeedRecallDeployment 判断是否需要回收扩张的资源
/*
- 注意：对于回收资源 最重要的逻辑是当前资源是否盈余 关键指标是tcp连接数的均值是否低于阈值
  只有tcp连接数低于阈值才能有最初的判断 --> 有回收资源的可能
  不需要回收返回false 需要回收返回True
  最后如果ch中 返回两个true才能回收 否则不回收
*/
func (tcpStrategy *TCPConnectStrategy) CheckIfNeedRecallDeployment() bool {
	if tcpStrategy.GetExtendDeploymentNum() == 0 {
		return false
	}

	// 否则当前有扩张的元素
	wgRecall.Add(strategiesRecallNum)
	go tcpStrategy.strategyFunctionBelowFloatRate()
	go tcpStrategy.strategyFunctionCountMeanSlopeRecall()
	wgRecall.Wait()

	vote := true
	for i := 0; i < strategiesRecallNum; i++ {
		vote = vote && <-chRecall
	}

	// 最终判断必须两个策略投票都是true 同时hasExtendDeploymentInFiveMinute 为false 表示5分钟内没有扩张过资源
	//fmt.Printf("!!!!!!!!!!!!!!!看看,GetHasExtendDeploymentInFiveMinute() %t\n", GetHasExtendDeploymentInFiveMinute())
	return vote && !GetHasExtendDeploymentInFiveMinute()
}

// checkIfCanExtend 检查是否允许/可以 扩容
func (tcpStrategy *TCPConnectStrategy) checkIfCanExtend() bool {
	if tcpStrategy.GetExtendDeploymentNum() >= maxScalingDeployments {
		// 如果超过当前的扩容数量大于了 deployment的最大伸缩量的情况的话 就直接拒绝了
		return false
	}

	// 否则deployment有可以扩容的空间 但是需要做额外的判断
	// 如果hasExtendDeploymentInFiveMinute 为false 即5分钟没有给deployment扩容 那么就可以扩容
	//fmt.Println("难道是这里！！=====", !GetHasExtendDeploymentInFiveMinute())
	return !GetHasExtendDeploymentInFiveMinute()
}

/*
  - strategyFunctionCountMeanSlope 计算斜率的平均变动情况
  - 每10秒计算一次 k1、k2、k3、k4、k5 6个区段 5个k值
    如果流量激增 那么tcp连接数就会逐渐增加 ，如果斜率的平均的增长速率大于一个阈值 那么就判断当前流量激增的非常迅速 需要马上扩容
    在实现上来说 其实是一个滑动窗口  在窗口内观察每一个时间间隔内的流量变动情况 也就是k的变动情况 然后求这个窗口内的流量增长速率

- TODO：滑动窗口应该有三个 分别为3分钟 5 分钟 10 分钟 然后全部都算一遍，现在优先实现3分钟的
*/
func (tcpStrategy *TCPConnectStrategy) strategyFunctionCountMeanSlope() {
	defer wgExtend.Done()
	// 如果funcEachTcpThreeMinute的长度大于2了就可以计算斜率了
	if len(funcEachTcpThreeMinute) >= 2 && !countSlopTimes {
		countSlopTimes = true
	}

	// 如果slopes容量满了 就要抛出一个元素
	if len(funcSlopesThreeMinute) >= funcSlopesThreeMinuteVolume {
		funcSlopesThreeMinute = funcSlopesThreeMinute[1:]
	}

	funcTcpSize := len(funcEachTcpThreeMinute)
	// 计算斜率
	if countSlopTimes {
		funcSlopesThreeMinute = append(
			funcSlopesThreeMinute,
			float32((funcEachTcpThreeMinute[funcTcpSize-1]-funcEachTcpThreeMinute[funcTcpSize-2])/TimeSet),
		)
	}

	// 计算窗口内的斜率的平均增速
	// 如果大于就要扩张 否则不用
	// 注意使用切片不能用迭代
	// 服务启动的前10秒内 funcSlopesThreeMinuteSize为0无法计算斜率
	sum := float32(0)
	funcSlopesThreeMinuteSize := len(funcSlopesThreeMinute)
	for i := 0; i < funcSlopesThreeMinuteSize; i++ {
		sum += funcSlopesThreeMinute[i]
	}

	if funcSlopesThreeMinuteSize == 0 {
		funcSlopesThreeMinuteSize = 1
	}

	// 计算窗口内斜率的平均变化程度
	res := int(sum)/funcSlopesThreeMinuteSize > averageGrowthRateThreshold
	log.Printf("平均流量增速 vote：%t\n", res)
	log.Println("三分钟内tcp连接增长速率： ", funcSlopesThreeMinute)
	chExtend <- res
}

/*
  - strategyFunctionExceedFloatRate 判断三分钟时间内流量是否超越了最大流量阈值
  - 限定最大流量阈值 maxFloatRate ，当流量超过maxFloatRate的时候直接投票 需要水平扩容容器
    需要注意：流量可能在一个时间内波动很大 ，所以需要阈值不能设置的太低 需要根据业务场景和业务时间来设定阈值
    如在晚上流量平缓上升 但是慢慢超过阈值 这个时候也是需要水平扩容 所以在超过阈值的情况下 增加时间纬度
    如果超过阈值的时间超过一段时间 那么就要水平扩容

  - 也是计算三分钟时间内流量是否一直超过阈值
    维护一个切片 切片记录一个窗口时间内的tcp连接数量 如果窗口时间内 3/4的tcp连接数量超过阈值 那么就删除
*/
func (tcpStrategy *TCPConnectStrategy) strategyFunctionExceedFloatRate() {
	defer wgExtend.Done()
	exceedNum := 0
	for i := 0; i < len(funcEachTcpThreeMinute); i++ {
		if funcEachTcpThreeMinute[i] > maxFloatRate {
			exceedNum++
		}
	}

	// 判断当前tcp连接数 是否超过了负载因子*窗口容量
	res := float32(exceedNum) > loadFactor*float32(len(funcEachTcpThreeMinute))
	log.Printf("流量最大负载 vote：%t\n", res)
	chExtend <- res
}

/*
*
  - strategyFunctionMaxMinGap 窗口时间内 最大值和最小值的差
  - 如果一段时间内tcp连接数呈现波动上升的话 CountMeanSlope策略就可能不适用了 尤其是当窗口范围增加的时候
    窗口内部的波动部分无法用来判断斜率的增益效果 因为斜率可能会为负值 所以就只能通过ExceedFloatRate 来判断了
    在这种情况下就要判断窗口内的tcp连接最大值和最值的差
*/
func (tcpStrategy *TCPConnectStrategy) strategyFunctionMaxMinGap() {
	defer wgExtend.Done()
	minTcpConnect := math.MaxInt32
	maxTcpConnect := 0

	// 维护窗口内的最大值和最小值
	for i := 0; i < len(funcEachTcpThreeMinute); i++ {
		if funcEachTcpThreeMinute[i] > maxTcpConnect {
			maxTcpConnect = funcEachTcpThreeMinute[i]
		}

		if funcEachTcpThreeMinute[i] < minTcpConnect {
			minTcpConnect = funcEachTcpThreeMinute[i]
		}
	}

	res := maxTcpConnect-minTcpConnect > 2*minTcpConnect
	log.Printf("流量最大差值 vote：%t\n", res)
	chExtend <- res
}

// recordTcpConnectionTreeMinute 计算三分钟时间内的tcp连接数
func (tcpStrategy *TCPConnectStrategy) recordTcpConnectionTreeMinute(tcpRes *resource.Tcp) {
	tcpConnectNum := tcpRes.TcpConnect.TcpNum

	// 窗口满了 就要先抛出元素
	if len(funcEachTcpThreeMinute) >= funcSingleTcpVolume {
		funcEachTcpThreeMinute = funcEachTcpThreeMinute[1:]
	}

	// 追加元素
	funcEachTcpThreeMinute = append(funcEachTcpThreeMinute, tcpConnectNum)
	log.Println("三分钟内tcp连接数： ", funcEachTcpThreeMinute)
}

// strategyFunctionBelowFloatRate 统计当前tcp的连接数是否可以低于阈值
func (tcpStrategy *TCPConnectStrategy) strategyFunctionBelowFloatRate() {
	defer wgRecall.Done()
	belowNum := 0
	for i := 0; i < len(funcEachTcpThreeMinute); i++ {
		if funcEachTcpThreeMinute[i] < maxFloatRate {
			belowNum++
		}
	}

	// 判断是否有3/4的时间 tcp连接数是小于阈值的 确实容器删除的条件还要严苛一点 负载因子应该要再大一点
	res := float32(belowNum) > loadFactor*float32(len(funcEachTcpThreeMinute))
	log.Printf("回收状态审验 计算阈值 vote: %t\n", res)
	chRecall <- res
}

// strategyFunctionCountSlop 统计当前tcp连接数变化的斜率
func (tcpStrategy *TCPConnectStrategy) strategyFunctionCountMeanSlopeRecall() {
	defer wgRecall.Done()
	sum := float32(0)

	for i := 0; i < len(funcSlopesThreeMinute); i++ {
		sum += funcSlopesThreeMinute[i]
	}

	// 统计斜率 希望斜率是递减的 希望斜率最后的平均变化情况是小于等于0的
	res := int(sum)/len(funcSlopesThreeMinute) <= 0
	log.Printf("回收状态审验 统计斜率 vote: %t\n", res)
	chRecall <- res
}
