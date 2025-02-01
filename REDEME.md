# EC2 自动备份与清理脚本

该项目使用 [AWS SDK for Go V2](https://github.com/aws/aws-sdk-go-v2) 来自动创建指定 EC2 实例的 AMI 备份，并清理过期的旧备份。

## 功能简介

1. **获取指定实例的卷信息**（系统盘 + 数据盘）。
2. **按当前实例卷设置创建 AMI**，并打上时间戳和自定义标签，支持指定最大保留数。
3. **定期清理旧镜像**，包括删除对应的快照。

## 快速开始

### 1. 克隆代码

```bash
git clone https://github.com/HsiangSun/awsAutoBackup.git
cd awsAutoBackup
go mod tidy
./build
```


### 2. 配置环境
```shell
//set AWS_ACCESS_KEY_ID=YOUR_ACCESS_KEY
//set AWS_SECRET_ACCESS_KEY=YOUR_SECRET_KEY
//set AWS_REGION=your-region
//set TARGET_INSTANCE=target-instance-id
```
