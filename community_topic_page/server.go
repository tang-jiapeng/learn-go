package main

import (
	"community_topic_page/controller"
	"community_topic_page/repository"
	"gopkg.in/gin-gonic/gin.v1"
	"os"
)

func main() {
	// 初始化数据索引
	if err := Init("./data/"); err != nil {
		os.Exit(-1)
	}
	// 初始化引擎配置
	r := gin.Default()
	// 构建路由
	r.GET("/community/page/get/:id", func(c *gin.Context) {
		topicId := c.Param("id")
		data := controller.QueryPageInfo(topicId)
		c.JSON(200, data)
	})
	r.POST("/community/post/do", func(c *gin.Context) {
		topicId, _ := c.GetPostForm("topic_id")
		content, _ := c.GetPostForm("content")
		data := controller.PublishPost(topicId, content)
		c.JSON(200, data)
	})
	// 启动服务
	err := r.Run()
	if err != nil {
		return
	}
}

func Init(filePath string) error {
	if err := repository.Init(filePath); err != nil {
		return err
	}
	return nil
}
