package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"math/big"

	//"demo/block"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

//==========================================区块链操作====================================================================
type Block struct {
	Index         int64
	TimeStamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int64 //This line is new
}
type Blockchain struct {
	blocks []*Block
}

//保存Hash结果在当前块的Hash中
func (b *Block) setHash() {
	b.Nonce = 0
	pow := NewProofOfWork(b)
	nonce, hash := pow.Run()
	b.Hash = hash[:] //保存Hash结果在当前块的Hash中
	b.Nonce = nonce
}

//构建Block新对象
func NewBlock(index int64, data, prevBlockHash []byte) *Block {
	block := &Block{index, time.Now().Unix(), data,
		prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

//添加新区块到链
func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(prevBlock.Index+1, []byte(data), prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}

//创建创世区块
func NewGenesisBlock() *Block {
	return NewBlock(0, []byte("first block"), []byte{})
}

//将创世区块入链
func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{NewGenesisBlock()}}
}

//==========================================工作量证明====================================================================
//困难程度
const targetBit = 8

//工作量证明的结构体
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

//新建工作量证明的结构体
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBit))
	pow := &ProofOfWork{b, target}
	return pow
}

//工作量函数
func (pow *ProofOfWork) Run() (int64, []byte) {
	var hashInt big.Int
	var hash [32]byte
	var nonce int64 = 0
	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)
	for {
		dataBytes := pow.prepareData(nonce) //获取准备的数据
		hash = sha256.Sum256(dataBytes)     //对数据进行Hash
		hashInt.SetBytes(hash[:])
		fmt.Printf("hash: \r%x", hash)
		if pow.target.Cmp(&hashInt) == 1 { //对比hash值
			break
		}
		nonce++ //充当计数器，同时在循环结束后也是符合要求的值

	}
	fmt.Printf("\n碰撞次数: %d\n", nonce)
	return int64(nonce), hash[:]
}

//准备数据
func (pow *ProofOfWork) prepareData(nonce int64) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			IntToHex(pow.block.TimeStamp),
			IntToHex(int64(targetBit)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}

//============================================工具函数===================================================================
//时间戳转换成日期
func UnixToTime(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 15:04:05")
}

//[]byte转换成string
func ByteToString(bytess []byte) string {
	t := string(bytess)
	return t
}
func ByteToHex(ten []byte) string {
	return hex.EncodeToString(ten)
}
func IntToHex(data int64) []byte {
	buffer := new(bytes.Buffer) // 新建一个buffer
	err := binary.Write(buffer, binary.BigEndian, data)
	if nil != err {
		log.Panicf("int to []byte failed! %v\n", err)
	}
	return buffer.Bytes()
}

//=============================================主函数====================================================================
func main() {
	router := gin.Default()
	router.SetFuncMap(template.FuncMap{
		"UnixToTime":   UnixToTime,
		"ByteToString": ByteToString,
		"ByteToHex":    ByteToHex,
	})
	router.LoadHTMLGlob("templates/*")
	//静态资源
	router.Static("/static", "./static")
	//===============================================哈希================================================================
	//哈希页面展示
	router.GET("/hash", func(c *gin.Context) {
		c.HTML(http.StatusOK, "hash.html", gin.H{
			"title": "hash",
		})
	})
	//哈希页面计算哈希
	router.POST("/hashComputeHash", func(c *gin.Context) {
		data := []byte(c.PostForm("Data"))
		hash := sha256.Sum256(data)
		c.HTML(http.StatusOK, "hash.html", gin.H{
			"title": "hash",
			"Data":  data,
			"Hash":  hash[:],
		})
	})
	//============================================区块===================================================================
	//区块界面用的区块
	blockBlock := NewBlock(0, []byte("first block"), []byte{})
	//区块首页展示
	router.GET("/block", func(c *gin.Context) {
		c.HTML(http.StatusOK, "block.html", gin.H{
			"title": "block",
			"block": &blockBlock,
		})
	})
	//区块页面修改数据即挖矿
	router.POST("/editBlock", func(c *gin.Context) {
		data := []byte(c.PostForm("Data"))
		blockBlock.Data = data
		blockBlock.setHash()
		c.Redirect(http.StatusMovedPermanently, "/block")
	})
	//=========================================区块链====================================================================
	//区块链界面用的区块
	blockchainBlockChain := NewBlockchain()
	blockchainBlockChain.AddBlock("data1")
	blockchainBlockChain.AddBlock("data2")
	//区块链首页展示
	router.GET("/blockchain", func(c *gin.Context) {
		c.HTML(http.StatusOK, "blockchain.html", gin.H{
			"title":                "blockchain",
			"blockchainBlockChain": &blockchainBlockChain.blocks,
		})
	})
	//接受增加区块的请求
	router.POST("/addBlockChain", func(c *gin.Context) {
		data := c.PostForm("data")
		blockchainBlockChain.AddBlock(data)
		c.Redirect(http.StatusMovedPermanently, "/blockchain")
	})
	//接受修改区块的请求
	router.POST("/editBlockChain", func(c *gin.Context) {
		Index, _ := strconv.Atoi(c.Query("Index"))
		data := []byte(c.PostForm("Data"))
		if Index != 0 {
			blockchainBlockChain.blocks[Index].PrevBlockHash = blockchainBlockChain.blocks[Index-1].Hash
		}
		block := blockchainBlockChain.blocks[Index]
		block.Data = data
		block.setHash()
		blockchainBlockChain.blocks[Index] = block
		c.Redirect(http.StatusMovedPermanently, "/blockchain")
	})
	//=========================================分布式====================================================================
	bc1 := NewBlockchain()
	bc1.AddBlock("data1")
	bc1.AddBlock("data2")
	bc1.AddBlock("data3")
	bc1.AddBlock("data4")
	bc1.AddBlock("data5")
	bc2 := Blockchain{}
	bc3 := Blockchain{}
	for i := 0; i < len(bc1.blocks); i++ {
		block := &Block{}
		block.Index = bc1.blocks[i].Index
		block.Data = bc1.blocks[i].Data
		block.PrevBlockHash = bc1.blocks[i].PrevBlockHash
		block.Hash = bc1.blocks[i].Hash
		block.Nonce = bc1.blocks[i].Nonce
		block.TimeStamp = bc1.blocks[i].TimeStamp
		bc2.blocks = append(bc2.blocks, block)
		bc3.blocks = append(bc3.blocks, block)
	}
	//首页展示
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index1.html", gin.H{
			"title":           "blockchains",
			"BlockChainList1": &bc1.blocks,
			"BlockChainList2": &bc2.blocks,
			"BlockChainList3": &bc3.blocks,
		})
	})
	//区块链1接受修改区块的请求
	router.POST("/editBlockChain1", func(c *gin.Context) {
		Index, _ := strconv.Atoi(c.Query("Index"))
		data := []byte(c.PostForm("Data"))
		if Index != 0 {
			bc1.blocks[Index].PrevBlockHash = bc1.blocks[Index-1].Hash
		}
		block := bc1.blocks[Index]
		block.Data = data
		block.setHash()
		bc1.blocks[Index] = block
		c.Redirect(http.StatusMovedPermanently, "/")
	})
	//区块链2接受修改区块的请求
	router.POST("/editBlockChain2", func(c *gin.Context) {
		Index, _ := strconv.Atoi(c.Query("Index"))
		data := []byte(c.PostForm("Data"))
		if Index != 0 {
			bc2.blocks[Index].PrevBlockHash = bc2.blocks[Index-1].Hash
		}
		block := bc2.blocks[Index]
		block.Data = data
		block.setHash()
		bc2.blocks[Index] = block
		c.Redirect(http.StatusMovedPermanently, "/")
	})
	//区块链3接受修改区块的请求
	router.POST("/editBlockChain3", func(c *gin.Context) {
		Index, _ := strconv.Atoi(c.Query("Index"))
		data := []byte(c.PostForm("Data"))
		if Index != 0 {
			bc3.blocks[Index].PrevBlockHash = bc3.blocks[Index-1].Hash
		}
		block := bc3.blocks[Index]
		block.Data = data
		block.setHash()
		bc3.blocks[Index] = block
		c.Redirect(http.StatusMovedPermanently, "/")
	})
	//==================================================================================================================
	router.Run()

}
