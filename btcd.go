package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/baiguwang/btc2omg/btcd/treasury"
	"github.com/btcsuite/baiguwang/btc2omg/omgd"
	"github.com/btcsuite/btcd/btc2omg/btcd/blockchain"
	"github.com/btcsuite/btcd/btc2omg/btcd/database"
	cross "github.com/btcsuite/btcd/btc2omg/btcd/wire"
	"github.com/btcsuite/btcd/btc2omg/btcd/wire/common"
	"github.com/btcsuite/btcd/btc2omg/omega/token"
	chaincfg "github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 添加调试模式开关和日志辅助函数
var debugMode bool = true

func logDebug(format string, args ...interface{}) {
	if debugMode {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// 联盟链交易数据结构
type AllianceL2Data struct {
	Hash   chainhash.Hash
	Height int32
	Txs    []*common.MsgXrossL2
}

// 定义结构以解析 Node.js 发送的数据
type NodeMessage struct {
	NotifyType string `json:"notifyType"`
	Height     int    `json:"height"`
	TxHash     string `json:"hash"`
	PID        string `json:"pid"`
	OID        string `json:"oid"`
	CID        string `json:"cid"`
}

// 定义结构以解析 RESTful 返回的交易数据
type RestTx struct {
	Inputs  []RestTxInput  `json:"inputs"`
	Outputs []RestTxOutput `json:"outputs"`
}

type RestTxInput struct {
	MintTxid  string `json:"spentTxid"`
	MintIndex uint32 `json:"mintIndex"`
	Sequence  uint32 `json:"sequence"`
	Script    string `json:"script"`
	Coinbase  bool   `json:"coinbase"`
	Address   string `json:"address"`
	Value     int64  `json:"value"`
}

type RestTxOutput struct {
	Value        int64        `json:"value"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}

type ScriptPubKey struct {
	Hex string `json:"hex"`
}

// 全局变量：已处理的哈希
var processedHashes = struct {
	sync.Mutex
	hashes []string
}{hashes: make([]string, 0)}

// 全局变量：区块链通知回调和通道
var (
	notificationCallback blockchain.NotificationCallback
	chainHeight          int32 = 0
)

// 检查哈希是否已经处理过
func isHashProcessed(hash string) bool {
	processedHashes.Lock()
	defer processedHashes.Unlock()

	for _, h := range processedHashes.hashes {
		if h == hash {
			return true
		}
	}

	processedHashes.hashes = append(processedHashes.hashes, hash)
	if len(processedHashes.hashes) > 100 {
		processedHashes.hashes = processedHashes.hashes[1:]
	}

	return false
}

// 读取最新高度
func readLatestHeight() int {
	file, err := os.Open("latest_height.txt")
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		height, err := strconv.Atoi(scanner.Text())
		if err == nil {
			return height
		}
	}
	return 0
}

// 更新最新高度
func updateLatestHeight(height int) {
	file, err := os.Create("latest_height.txt")
	if err != nil {
		fmt.Println("更新高度文件错误:", err)
		return
	}
	defer file.Close()

	_, _ = file.WriteString(fmt.Sprintf("%d", height))
	chainHeight = int32(height)
}

// revHex 将大端格式的十六进制哈希转换为小端格式
func revHex(data string) string {
	bytes, err := hex.DecodeString(data)
	if err != nil {
		fmt.Println("解码十六进制失败:", err)
		return ""
	}

	reversed := make([]byte, len(bytes))
	for i := 0; i < len(bytes); i++ {
		reversed[i] = bytes[len(bytes)-1-i]
	}

	return hex.EncodeToString(reversed)
}

// 查询 RESTful 接口获取交易详情
func fetchTransaction(txHash string) (*RestTx, error) {
	url := fmt.Sprintf("https://explorer.gamegold.xin/public/tx/%s/coins", txHash)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("获取交易失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("非200状态码: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	var data struct {
		Code   int    `json:"code"`
		Result RestTx `json:"result"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v", err)
	}

	if data.Code != 0 {
		return nil, fmt.Errorf("RESTful API返回错误码: %d", data.Code)
	}

	return &data.Result, nil
}

// 将 RestTx 转换为 btcsuite 的 MsgTx
func convertToMsgTx(tx *RestTx) (*wire.MsgTx, error) {
	msgTx := &wire.MsgTx{
		Version:  1,
		TxIn:     []*wire.TxIn{},
		TxOut:    []*wire.TxOut{},
		LockTime: 0,
	}

	for _, input := range tx.Inputs {
		hash, err := chainhash.NewHashFromStr(input.MintTxid)
		if err != nil {
			return nil, fmt.Errorf("无效哈希: %v", err)
		}

		outPoint := wire.OutPoint{
			Hash:  *hash,
			Index: input.MintIndex,
		}

		msgTx.TxIn = append(msgTx.TxIn, &wire.TxIn{
			PreviousOutPoint: outPoint,
			SignatureScript:  []byte(input.Script),
			Sequence:         input.Sequence,
		})
	}

	for _, output := range tx.Outputs {
		pkScript, err := hex.DecodeString(output.ScriptPubKey.Hex)
		if err != nil {
			return nil, fmt.Errorf("无效的PkScript: %v", err)
		}

		msgTx.TxOut = append(msgTx.TxOut, &wire.TxOut{
			Value:    output.Value,
			PkScript: pkScript,
		})
	}

	return msgTx, nil
}

// 处理来自Node.js的消息
// 处理来自Node.js的消息
func handleNodeMessage(message string) {
	logDebug("收到原始消息: %s", message)
	logDebug("回调状态: %v", notificationCallback != nil)

	// 检查消息是否为空或格式不正确
	message = strings.TrimSpace(message)
	if message == "" {
		logDebug("忽略空消息")
		return
	}

	var nodeMsg NodeMessage
	if err := json.Unmarshal([]byte(message), &nodeMsg); err != nil {
		fmt.Println("消息解析失败:", err)
		logDebug("消息内容: %s", message)
		return
	}

	// 检查基本字段
	if nodeMsg.TxHash == "" {
		logDebug("忽略没有交易哈希的消息: %+v", nodeMsg)
		return
	}

	// 根据消息类型处理哈希格式
	if nodeMsg.NotifyType == "prop.exchange" {
		nodeMsg.TxHash = revHex(nodeMsg.TxHash)
		logDebug("转换哈希格式: %s", nodeMsg.TxHash)
	}

	logDebug("解析后的消息: %+v", nodeMsg)

	// 检查是否已处理过该哈希
	if isHashProcessed(nodeMsg.TxHash) {
		logDebug("哈希已处理: %s", nodeMsg.TxHash)
		return
	}

	// 获取交易详情
	tx, err := fetchTransaction(nodeMsg.TxHash)
	if err != nil {
		fmt.Printf("获取交易 %s 失败: %v\n", nodeMsg.TxHash, err)
		return
	}
	logDebug("成功获取交易详情，输入: %d, 输出: %d", len(tx.Inputs), len(tx.Outputs))

	// 转换为MsgTx格式
	msgTx, err := convertToMsgTx(tx)
	if err != nil {
		fmt.Printf("转换为MsgTx失败: %v\n", err)
		return
	}
	logDebug("成功转换为MsgTx格式")

	// 创建跨链交易数据
	hash, err := chainhash.NewHashFromStr(nodeMsg.TxHash)
	if err != nil {
		fmt.Printf("创建哈希对象失败: %v\n", err)
		return
	}

	xrossL2Txs := make([]*common.MsgXrossL2, 0)

	for i, output := range tx.Outputs {
		// 判断是否符合跨链条件（这里的条件可能需要调整）
		if nodeMsg.NotifyType == "prop.exchange" ||
			nodeMsg.NotifyType == "block.connect" ||
			nodeMsg.NotifyType == "sys.rescan" {

			pkScript, _ := hex.DecodeString(output.ScriptPubKey.Hex)

			// 创建跨链交易数据
			outPoint := wire.OutPoint{
				Hash:  *hash,
				Index: uint32(i),
			}

			xrossL2Txs = append(xrossL2Txs, common.XrossL2(&outPoint, output.Value, pkScript))
		}
	}

	// 处理交易输出
	for i, txOut := range msgTx.TxOut {
		logDebug("检查输出 #%d: 值=%d, 脚本长度=%d", i, txOut.Value, len(txOut.PkScript))

		// 放宽交易条件，只要有价值的输出都考虑为跨链交易
		if txOut.Value > 0 {
			outPoint := wire.OutPoint{
				Hash:  *hash,
				Index: uint32(i),
			}

			// 创建跨链交易
			xrossL2 := common.XrossL2(&outPoint, txOut.Value, txOut.PkScript)
			if xrossL2 != nil {
				xrossL2Txs = append(xrossL2Txs, xrossL2)
				logDebug("创建跨链交易 #%d: 输出点=%s:%d, 值=%d",
					len(xrossL2Txs), hash.String(), i, txOut.Value)
			} else {
				logDebug("创建跨链交易失败，XrossL2函数返回nil")
			}
		}
	}

	// 如果找到跨链交易且回调已注册，发送通知
	logDebug("找到 %d 个跨链交易, 回调状态: %v", len(xrossL2Txs), notificationCallback != nil)

	if len(xrossL2Txs) > 0 {
		if notificationCallback != nil {
			logDebug("准备调用回调函数...")
			data := &AllianceL2Data{
				Hash:   *hash,
				Height: int32(nodeMsg.Height),
				Txs:    xrossL2Txs,
			}

			// 发送跨链交易通知
			//notification := &blockchain.Notification{
			//	Type: blockchain.NTBTCTxConnected,
			//	Data: data,
			//}
			omgd.Server.Chain.SendNotification(blockchain.NTBTCTxConnected, data)

			// 尝试调用回调并捕获可能的异常
			func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Println("回调函数执行异常:", r)
						logDebug("异常详情: %v", r)
					}
				}()
				//notificationCallback(notification)
				logDebug("回调函数执行完成")
			}()

			fmt.Printf("已发送跨链交易通知: %s, 包含 %d 个交易\n",
				hash.String(), len(xrossL2Txs))
		} else {
			fmt.Println("警告: 找到跨链交易但回调未注册")
		}
	} else {
		logDebug("没有找到跨链交易，不调用回调")
	}

	// 打印交易信息
	logDebug("处理的MsgTx: 版本=%d, 锁定时间=%d, 输入=%d, 输出=%d",
		msgTx.Version, msgTx.LockTime, len(msgTx.TxIn), len(msgTx.TxOut))

	// 更新处理高度
	updateLatestHeight(nodeMsg.Height)
	logDebug("更新处理高度为: %d", nodeMsg.Height)
}

// Connect - 联盟链连接函数，专注于回调处理
func Connect(subscriber blockchain.NotificationCallback, q chan interface{}) {
	// 直接保存订阅者回调
	notificationCallback = subscriber

	logDebug("Connect函数启动，回调状态: %v", notificationCallback != nil)

	// 注册区块链通知处理
	if omgd.Server != nil && omgd.Server.Chain != nil {
		omgd.Server.Chain.Subscribe(func(notification *blockchain.Notification) {
			// 处理来自区块链的通知
			logDebug("收到区块链通知: %v", notification.Type)

			// 根据通知类型处理
			switch notification.Type {
			case blockchain.NTBTCTxConnected:
				omgd.Server.Db.Update(func(dbtx database.Tx) error {
					tx := notification.Data.(*AllianceL2Data)

					bucket := dbtx.Metadata().Bucket([]byte(common.INCOMINGPOOL))
					//b2 := dbtx.Metadata().Bucket([]byte(common.BTCL2POOL))
					xchain := &cross.XchainData{
						ChainID:   common.BTCCHAINID,
						Hash:      *treasury.Bhash2l2hash(tx.Hash),
						Height:    tx.Height,
						Txs:       []*cross.MsgXrossL2{},
						Finalized: 0,
					}
					//xchain2 := &cross.XchainData{
					//	ChainID:   common.BTCCHAINID,
					//	Hash:      *treasury.Bhash2l2hash(tx.Hash),
					//	Height:    tx.Height,
					//	Txs:       []*cross.MsgXrossL2{},
					//	Finalized: 0,
					//}

					for _, t := range tx.Txs {
						tk := token.Token{
							TokenType: common.BTCCHAINID << 40,
							Value:     &token.NumToken{Val: t.Value},
							Rights:    nil,
						}

						//scp, ok := treasury.RecoverBTCScript(dbtx, t.PkScript)
						//if scp == nil {
						//	continue
						//}

						s := &cross.MsgXrossL2{
							Utxo: cross.OutPoint{
								Hash:  *treasury.Bhash2l2hash(t.Utxo.Hash),
								Index: t.Utxo.Index,
							},
							Txo: cross.TxOut{
								PkScript: t.PkScript,
							},
						}
						s.Txo.Token = tk

						//if ok {
						xchain.Txs = append(xchain.Txs, s)
						//} else {
						//	xchain2.Txs = append(xchain2.Txs, s)
						//}
					}

					//if len(xchain.Txs) > 0 { // confirmed txs of xfer BTC to BOVM
					var k [36]byte
					copy(k[:], xchain.Hash[:])
					common.LittleEndian.PutUint32(k[32:], uint32(xchain.ChainID|cross.CrossChainFalg))
					bucket.Put(k[:], xchain.Serialize())
					//}
					//if len(xchain2.Txs) > 0 { // unconfirmed txs of xfer BTC to BOVM
					//	var k [36]byte
					//	copy(k[:], xchain2.Hash[:])
					//	common.LittleEndian.PutUint32(k[32:], uint32(xchain2.ChainID|cross.CrossChainFalg))
					//	b2.Put(k[:], xchain2.Serialize())
					//}

					return nil
				})
			case blockchain.NTBlockDisconnected:
				logDebug("处理区块断开通知")
				// 处理断开通知的逻辑...
			}

			// 如果有订阅者回调，也传递通知
			if notificationCallback != nil {
				notificationCallback(notification)
			}
		})
		logDebug("已成功注册区块链通知处理")
	} else {
		fmt.Println("警告: omgd.Server 或 omgd.Server.Chain 为空，无法注册区块链通知")
	}

	fmt.Println("联盟链连接函数初始化完成")
}

// 与服务器建立连接并处理消息
func startServerConnection() {
	logDebug("启动服务器连接协程...")

	go func() {
		for {
			logDebug("正在连接联盟链服务器...")
			conn, err := net.Dial("tcp", "localhost:2011")
			if err != nil {
				fmt.Println("连接服务器错误:", err)
				time.Sleep(5 * time.Second)
				continue
			}

			// 发送初始化命令，从上次处理的高度开始
			startHeight := readLatestHeight()
			fmt.Fprintf(conn, "RESCAN %d\n", startHeight)
			logDebug("请求从高度 %d 开始扫描", startHeight)

			reader := bufio.NewReader(conn)
			for {
				message, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("连接断开，正在重连...")
					break
				}

				// 处理收到的消息
				handleNodeMessage(message)
			}

			conn.Close()
			time.Sleep(5 * time.Second)
		}
	}()

	fmt.Println("服务器连接协程已启动")
}

// GetChainHeight 返回当前区块链高度
func GetChainHeight() int32 {
	return chainHeight
}

func main() {
	// 获取中断信号
	interrupt := interruptListener()

	// 构造OMGD
	err := omgd.Construct(&chaincfg.MainNetParams)
	if err != nil {
		fmt.Println("构建OMGD失败:", err)
		return
	}

	//params := omgd.Server.Chain.ChainParams
	//encodeaddr := "1JKQqXbD5wpDGyGtmmsEypVjeG6GVxaKXE"
	//addr, _ := btcutil.DecodeAddress(encodeaddr, params)
	//
	//var h [20]byte
	//copy(h[:], addr.EncodeAddress())
	//
	//witnessadress, _, _ := treasury.Get75pctMSScript(h)
	//if witnessadress == "" {
	//}

	//// 创建通道（这里使用interface{}类型以兼容不同数据结构）
	//q := make(chan interface{}, 1000)
	//
	//// 创建独立的通知处理函数
	//notificationHandler := func(notification *blockchain.Notification) {
	//	fmt.Printf("接收到通知: 类型=%v\n", notification.Type)
	//	// 发送到队列供后续处理
	//	select {
	//	case q <- notification:
	//		fmt.Println("通知已加入队列")
	//	default:
	//		fmt.Println("队列已满，通知丢弃")
	//	}
	//}
	//
	//Connect(notificationHandler, q)
	//
	//// 2. 再启动服务器连接
	//startServerConnection()
	//
	//// 添加处理队列的协程
	//go func() {
	//	for {
	//		select {
	//		case notification, ok := <-q:
	//			if !ok {
	//				return
	//			}
	//			// 处理从队列取出的通知
	//			fmt.Printf("处理队列中的通知: %v\n", notification)
	//		}
	//	}
	//}()

	// 启动OMGD Layer 2
	shutdownChannel := make(chan struct{})
	err = omgd.Start(shutdownChannel)
	if err != nil {
		fmt.Println("启动OMGD失败:", err)
		return
	}

	// 等待中断信号
	<-interrupt

	// 关闭
	close(shutdownChannel)
	omgd.Stop()
}
