# 草田链节点部署

> #### 数据节点需要完成1,2两步.
> #### 加入验证节点需要完成所有步骤.
###### <font color='red'> 命令中使用到的私钥账户均为测试开源账户,请勿使用于生产环境!!! </font>
 
---

1. 下载二进制执行文件

   - Windows 
   [geth.exe](https://github.com/qydata/go-ctereum/releases/download/v1.10.26-stable/geth.exe)
   - Linux
   [geth](https://github.com/qydata/go-ctereum/releases/download/v1.10.26-stable/geth)
2. 运行程序(修改区块链文件夹和端口)

    设置环境变量
    ```dotenv
    export CHAIN_DIR=/data/ctnode
    export CHAIN_PORT=17474
    export CHAIN_HTTP_PORT=7474
    ```

    ```shell
    ./geth  \
    --datadir ${CHAIN_DIR}   \
    --ipcpath ${CHAIN_DIR}/geth.ipc \
    --port ${CHAIN_PORT} \
    --http \
    --http.api "eth,net,engine,admin,web3,debug,txpool" \
    --http.port ${CHAIN_HTTP_PORT} \
    --http.addr=0.0.0.0 \
    --allow-insecure-unlock \
    --syncmode full \
    --gcmode=archive  \
     > node1.out 2>&1 &
    ``` 
    查看节点数据同步和节点连接数状况
   
    首先查看节点数, 如果节点没有连接到其它节点,则需要手动添加节点,来进行数据同步,默认情况下,节点会自动发现和设置节点
   
    ```log
    > net
    {
        listening: true,
        peerCount: 2,
        version: "27",
        getListening: function(callback),
        getPeerCount: function(callback),
        getVersion: function(callback)
    }
    ```
    这里可以看到节点数是2, 证明节点是正常发现的.   

    查看节点同步状况
    ```log
    eth.syncing
   false
   > eth.blockNumber
    5160625
    ```
    通过查看节点的区块数和同步状态,来对比[草田链浏览器](http://ctblock.cn) 来确认节点是否已经同步为最新的状态节点同步到最新状态既可以使用节点来查询和进行上链操作
    
>数据节点到此结束

3. 在此创建节点验证账户
[链上工具](https://wallet.ctblock.cn/)
   * <font color='red'> 需要牢记账户私钥 </font>
    

4. 进入前面搭建的节点的节点控制台
   1. 导入账户
        ```javascript
        personal.importRawKey("2e0834786285daccd064ca17f1654f67b4aef298acbb82cef9ec422fb4975622", "ttttt")
        ```
   2. 解锁账户
      ```javascript
      personal.unlockAccount("0x123463a4b065722e99115d6c222f267d9cabb524", "ttttt", 0)
      ```
   3. 设置手续费账户
        ```javascript
        miner.setEtherbase("0x123463a4b065722e99115d6c222f267d9cabb524")
        ```
   4. 开启验证人工作(并未真的开始验证, 只是准备工作)
        ```javascript
        miner.start()
        ```

5. 质押草田分参与验证
   1. 购买草田分
   2. 质押草田分
      
      通过在以下界面质押指定数额的草田分来参与挖矿
      
   3. 查看节点日志, 观察是否参与挖矿(这里会有最长五分钟延迟).出现类似以下日志, 则证明节点已经参与验证数据
    ```log
    INFO [03-08|15:51:20.000] Successfully sealed new block            number=5,160,667 sealhash=07d848..2a4dd8 hash=d36bad..d434a1 elapsed=4.999s
    INFO [03-08|15:51:20.000] 🔗 block reached canonical chain          number=5,160,660 hash=b75f69..a1a4ca
    INFO [03-08|15:51:20.001] 🔨 mined potential block                  number=5,160,667 hash=d36bad..d434a1
    ```

