package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"log"
	"os"
	"sort"
	"time"
)

const (
	maxBackups     = 1    // 最大保留备份数
	noRebootBackup = true // 创建镜像时不重启实例
	backupTag      = "DailyBackup"
)

func main() {
	//set AWS_ACCESS_KEY_ID=YOUR_ACCESS_KEY
	//set AWS_SECRET_ACCESS_KEY=YOUR_SECRET_KEY
	//set AWS_REGION=your-region
	//set TARGET_INSTANCE=target-instance-id

	instanceID := os.Getenv("TARGET_INSTANCE")
	if instanceID == "" {
		log.Fatalf("please set env TARGET_INSTANCE firstly")
	}

	// 加载 AWS 配置
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("无法加载 AWS 配置: %v", err)
	}

	// 创建 EC2 客户端
	client := ec2.NewFromConfig(cfg)

	// 创建镜像
	createAndCleanupBackups(client, instanceID)
	//cleanupOldImages(client)
}

// 创建镜像并清理旧备份
func createAndCleanupBackups(client *ec2.Client, instanceID string) {
	// 创建新镜像
	timestamp := time.Now().Format("20060102-1504")
	imageName := backupTag + "-" + timestamp

	// 创建镜像
	_, err := client.CreateImage(context.TODO(), &ec2.CreateImageInput{
		InstanceId:  aws.String(instanceID),
		Name:        aws.String(imageName),
		Description: aws.String("auto-backup-" + timestamp),
		NoReboot:    aws.Bool(noRebootBackup),
		//BlockDeviceMappings: blockDeviceMappings,
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeImage,
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(imageName),
					},
				},
			},
		},
	})

	if err != nil {
		log.Printf("创建镜像失败: %v", err)
		return
	}

	log.Printf("已创建镜像: %s", imageName)

	// 清理旧备份
	cleanupOldImages(client)
}

// 清理旧镜像
func cleanupOldImages(client *ec2.Client) {
	images, err := client.DescribeImages(context.TODO(), &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []string{backupTag + "*"},
			},
		},
		Owners: []string{"self"},
	})
	if err != nil {
		log.Printf("查询镜像失败: %v", err)
		return
	}

	// 按创建时间排序（旧到新）
	sort.Slice(images.Images, func(i, j int) bool {
		return *images.Images[i].CreationDate < *images.Images[j].CreationDate
	})

	log.Printf("Found %s images count = %d", backupTag, len(images.Images))

	// 删除超过保留数量的旧镜像
	if len(images.Images) > maxBackups {
		for _, image := range images.Images[:len(images.Images)-maxBackups] {
			// 删除镜像
			_, err := client.DeregisterImage(context.TODO(), &ec2.DeregisterImageInput{
				ImageId: image.ImageId,
			})
			if err != nil {
				log.Printf("删除镜像 %s 失败: %v", *image.Name, err)
				continue
			}

			// 删除关联的快照
			for _, bd := range image.BlockDeviceMappings {
				if bd.Ebs != nil && bd.Ebs.SnapshotId != nil {
					_, err := client.DeleteSnapshot(context.TODO(), &ec2.DeleteSnapshotInput{
						SnapshotId: bd.Ebs.SnapshotId,
					})
					if err != nil {
						log.Printf("删除快照 %s 失败: %v", *bd.Ebs.SnapshotId, err)
					}
				}
			}

			log.Printf("已删除旧镜像: %s", *image.Name)
		}
	}
}
