package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wujunwei928/parse-video/parser"
)

type HttpResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type DownloadRequest struct {
	FolderPath   string `json:"folderPath"` // 文件夹名称
	DownloadUrls []struct {
		URL      string `json:"url"`
		Filename string `json:"filename"`
	} `json:"downloadUrls"`
}

func main() {
	r := gin.Default()

	// 加载静态文件
	r.Static("/static", "./static") // 将 "/static" 路径映射到项目目录的 "./static" 文件夹

	// 加载 HTML 模板文件
	r.LoadHTMLGlob("templates/*") // 加载 templates 文件夹中的所有模板文件

	// 设置路由，返回 index.html 页面
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "在线短视频去水印解析", // 可以传递数据到模板中
		})
	})

	r.GET("/video/share/url/parse", func(c *gin.Context) {
		paramUrl := c.Query("url")
		parseRes, err := parser.ParseVideoShareUrlByRegexp(paramUrl)
		jsonRes := HttpResponse{
			Code: 200,
			Msg:  "解析成功",
			Data: parseRes,
		}
		if err != nil {
			jsonRes = HttpResponse{
				Code: 201,
				Msg:  err.Error(),
			}
		}

		c.JSON(http.StatusOK, jsonRes)
	})

	r.POST("/api/download", func(c *gin.Context) {
		var req DownloadRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		// 遍历下载 URL
		for _, file := range req.DownloadUrls {
			// 下载文件
			err := downloadFile(file.URL, req.FolderPath, file.Filename)
			if err != nil {
				//c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download file"})
				//return
				log.Println("Failed to download file:", err)
			}
		}
		c.JSON(http.StatusOK, gin.H{"message": "All files downloaded successfully!"})
	})

	r.GET("/video/id/parse", func(c *gin.Context) {
		videoId := c.Query("video_id")
		source := c.Query("source")

		parseRes, err := parser.ParseVideoId(source, videoId)
		jsonRes := HttpResponse{
			Code: 200,
			Msg:  "解析成功",
			Data: parseRes,
		}
		if err != nil {
			jsonRes = HttpResponse{
				Code: 201,
				Msg:  err.Error(),
			}
		}

		c.JSON(200, jsonRes)
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器 (设置 5 秒的超时时间)
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")

}

// downloadFile 下载文件并保存到指定路径
func downloadFile(url string, fileFolder, filepath string) error {
	fileFolder = path.Join("D:\\带货短视频\\", fileFolder)
	EnsureDirExists(fileFolder)
	// 构造完整文件路径
	fullPath := path.Join(fileFolder, filepath)

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	out, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, response.Body)
	return err
}

// 判断目录是否存在，如果不存在则创建
func EnsureDirExists(dirPath string) (error, bool) {
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return err, false
		}
	} else if err != nil {
		return err, false
	} else if !info.IsDir() {
		return err, false
	} else {
		return nil, true
	}

	return nil, false
}
