# fuploadServer
## 一、说明
这个demo的目标是提供一个能断点上传大文件的组件，只要client端按照协商好的格式，上传数据，fuploadServer就能提供正确的上传服务。

## 二、运行example中的例子  
 准备好一个名字为test.mp4的文件，放在/home 下。  
 1、运行fuploaderServer   
   nohup ./fuploadServer -c ./conf/fupload.yml &   
  
 2、运行业务服务器  
   nohup ./businessServer &  
 
 3、运行client端  
   ./client   
  
  可以看到会把/home/test.mp4上传到上传组件所在机器上的/tmp/business/目录下。

## 三、使用 
#### 1、部署上传组件  
go get http:github.com/gmluckly/fuploadServer  
go build   
配置conf中的fupload.yaml文件，
bServer项为与业务服务器交互的信息配置，配置token和业务服务回调url  
启动：  
    nohup/fupload_server -c ./conf/fupload.yaml &   
    
也可以使用Dockerfile打包成镜像，在docker中运行  
docker build -t fuploadserver:v1.0.0 .  
docker run --name fupload -d -p8090:8090 fuploadserver:v1.0.0


#### 2、服务开发对接
当我们以组件的形式提供服务的时候，我们把角色分成三个：  
 (1)客户端(client) ,上传文件的角色   
 (2)业务服务器(business server)  
 (3)上传组件(fuploadServer)  

 流程：   
 ![image](https://github.com/gmluckly/fuploadServer/blob/master/fupload.jpg)
 
 #### 1、client端申请上传。
 ##### 1.1 client 与business server交互，申请上传，具体协议自定义  
 ##### 1.2 business server收到client端的上传请求，转发到fupload server  
 request:  
 method: post   
 content-Type:application/json  
 body:  
 {  
 &nbsp; &nbsp; "token": string, &nbsp;&nbsp; //业务服务与上传组件约定的token认证  
 &nbsp; &nbsp;     "userId": string, &nbsp;&nbsp;//上传的用户id   
 &nbsp; &nbsp;     "storePath": string,  &nbsp;&nbsp;//最终存放的路径  
 &nbsp; &nbsp;     "fileName":  string,  &nbsp;&nbsp;//文件名称  
 &nbsp; &nbsp;    "fileSize": int64,  &nbsp;&nbsp;//文件大小  
 &nbsp; &nbsp;    "fileMd5":  string,  &nbsp;&nbsp;//文件的md5值  
 &nbsp; &nbsp;    "feedBackParameter": interface{}  &nbsp;&nbsp;//上传后的回调参数，业务服务器可以封装为任意值，上传组件不会解析  
 } 
 
reponse:  
 {  
    &nbsp;&nbsp; "retCode": string &nbsp;&nbsp;//0成功，其他失败  
    &nbsp;&nbsp; "status": string  &nbsp;&nbsp;//失败原因  
    &nbsp;&nbsp; "data":{  
        &nbsp;&nbsp;&nbsp;&nbsp; "taskId": int64 &nbsp;&nbsp;//上传id  
        &nbsp;&nbsp;&nbsp;&nbsp; "uploadUrl": string &nbsp;&nbsp;//上传文件块url  
        &nbsp;&nbsp;&nbsp;&nbsp; "blocksUrl": string &nbsp;&nbsp;//查询上传任务信息url，获取上传组件期待的文件块  
        &nbsp;&nbsp;&nbsp;&nbsp; "stateUrl":string &nbsp;&nbsp;//更新上传任务的url
    }  
 }  
 
 例如，curl命令：  
 curl -i -H 'content-type: application/json' -X POST -d '{"token":"xxxxx","userId":"123","storePath":"/tmp/fupload/","fileName":"hello.mp4","fileSize":"345000000","fileMd5":"04465d416b4eca345976021db8698f15"}'&nbsp;&nbsp;http://ip:port/api/upload/new/task
 
 #### 2、client端与上传组件交互，上传文件块和暂停上传等操作。  
 ##### 2.1 上传文件块
 http 请求，url为申请上传时返回的uploadUrl  
 request:  
 &nbsp;&nbsp;&nbsp;&nbsp;method: post,  
 &nbsp;&nbsp;&nbsp;&nbsp;content-Type: multiplepart/form-data  
 body:  
 &nbsp;&nbsp;&nbsp;&nbsp;{  
     &nbsp;&nbsp;&nbsp;&nbsp;&nbsp; partMd5: string &nbsp;&nbsp;//块MD5的值  
     &nbsp;&nbsp;&nbsp;&nbsp;&nbsp; startByte: int64 &nbsp;&nbsp;//文件块开始的位置(字节)  
     &nbsp;&nbsp;&nbsp;&nbsp;&nbsp; startByte: int64 &nbsp;&nbsp; //文件块结束的位置(字节)  
     &nbsp;&nbsp;&nbsp;&nbsp;&nbsp; file: binary &nbsp;&nbsp; //块内容  
 &nbsp;&nbsp;&nbsp;&nbsp;}  
 
 reponse:  
  {  
    &nbsp;&nbsp; "retCode": string &nbsp;&nbsp;//0成功，1上传组件提示需要某些块,其他失败    
    &nbsp;&nbsp; "status": string  &nbsp;&nbsp;//失败原因  
    &nbsp;&nbsp; "data":{  
        &nbsp;&nbsp;&nbsp;&nbsp; startByte: int64 &nbsp;&nbsp;//开始位置  
        &nbsp;&nbsp;&nbsp;&nbsp; startByte: int64 &nbsp;&nbsp;//结束位置   
    &nbsp;&nbsp;&nbsp;&nbsp; &nbsp;&nbsp;}  
 } 
 
 例如，curl命令：  
 
 curl -i -H 'Content-Type:multipart/form-data' -X POST -F 'startByte=0' -F 'endByte=1000' -F 'partMd5=a01610228fe998f515a72dd730294d87' -F 'file=/tmp/1.mp4.01' http://ip:port/api/upload/files/block?taskId=123   // uploadUrl
 
 
##### 2.2 更改上传状态(暂停或开始)
http 请求，url为申请上传时返回的stateUrl  
request:  
 &nbsp;&nbsp;&nbsp;&nbsp;method: post,  
 &nbsp;&nbsp;&nbsp;&nbsp;content-Type:application/json  
 body:  
 &nbsp;&nbsp;&nbsp;&nbsp;{  
     &nbsp;&nbsp;&nbsp;&nbsp;&nbsp; taskId: int64 &nbsp;&nbsp;//上传任务id  
     &nbsp;&nbsp;&nbsp;&nbsp;&nbsp; status: string &nbsp;&nbsp; //"pause"暂停，"upload" 上传，"cancel"取消  
 &nbsp;&nbsp;&nbsp;&nbsp;}  
 
 reponse:  
  {  
    &nbsp;&nbsp; "retCode": string &nbsp;&nbsp;//0成功，其他失败    
    &nbsp;&nbsp; "status": string  &nbsp;&nbsp;//失败原因  
    &nbsp;&nbsp; "data": string     
 } 
 
 例如，curl命令：  
 
 curl -i -H 'content-type: application/json' -X POST -d '{"taskId":12344,"status":"pause"}'&nbsp;&nbsp;http://ip:port/api/upload/files/state  //stateUrl  
 
 ##### 2.3 获取上传组件期待列表
http 请求，url为申请上传时返回的blocksUrl  
request:  
 &nbsp;&nbsp;&nbsp;&nbsp;method: get,  
 
 reponse:  
{  
    &nbsp;&nbsp;&nbsp; totalSize: int64 &nbsp;&nbsp;//上传任务的大小  
    &nbsp;&nbsp;&nbsp; status: string &nbsp;&nbsp; //上传状态 "pause"暂停，"upload" 上传，"cancel"取消  
    &nbsp;&nbsp;&nbsp; expectBlocks: {  
          &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;  &nbsp;&nbsp; startByte: int64  
           &nbsp;&nbsp;&nbsp;&nbsp;&nbsp; &nbsp;&nbsp; endByte: int64  
     &nbsp;&nbsp;&nbsp;&nbsp;&nbsp; }  
}  
 例如，curl命令：  
 
 curl http://ip:port/api/upload/files/blocks/info?taskId=123  //blocksUrl  
 
 
#### 3、上传回调(上传组件主动回调业务服务器，业务服务需要实现这个接口)
http 请求，url为配置项中定义的notifyUrl 
request:  
 &nbsp;&nbsp;&nbsp;&nbsp;method: post,  
 &nbsp;&nbsp;&nbsp;&nbsp;content-Type:application/json  
 
body:  
 &nbsp;&nbsp;&nbsp;&nbsp;{  
     &nbsp;&nbsp;&nbsp;&nbsp;&nbsp; retCode: string &nbsp;&nbsp;//0成功，其他失败    
     &nbsp;&nbsp;&nbsp;&nbsp;&nbsp; status: string &nbsp;&nbsp; //失败原因  
     &nbsp;&nbsp;&nbsp;&nbsp;&nbsp; data: "" &nbsp;&nbsp; //申请上传时的回调参数，只有retCode为0 时返回  
 &nbsp;&nbsp;&nbsp;&nbsp;}  
 
 reponse:  
 忽略  
 