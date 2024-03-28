package util

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"
)

// Download 异步带进度下载文件
func Download(task []string, fileFolder string) {
	//"https://media.w3.org/2010/05/sintel/trailer.mp4",
	//"http://www.w3school.com.cn/example/html5/mov_bbb.mp4",
	//"https://www.w3schools.com/html/movie.mp4",
	//"http://devimages.apple.com/iphone/samples/bipbop/bipbopall.m3u8",
	//"http://speedtest.tele2.net/100MB.zip",

	// 文件保存目录
	//fileFolder := "d:/"

	var wg sync.WaitGroup
	for _, addr := range task {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			filename, err := extractFilenameFromURL(addr)
			if err != nil {
				log.Printf("无法提取文件名: %v\n", err)
				return
			}
			// 文件的保存路径
			savePath := fileFolder + filename
			if err = downloadFile(addr, savePath); err != nil {
				log.Printf("下载 %s 时出错: %v\n", addr, err)
			}
		}(addr)
	}
	wg.Wait()
}

// Downloader 提供了读取进度的功能
type Downloader struct {
	io.Reader
	Total   int64
	Current int64
}

// Read 实现了io.Reader接口，用于读取数据并更新进度
func (d *Downloader) Read(p []byte) (n int, err error) {
	n, err = d.Reader.Read(p)
	d.Current += int64(n)
	return
}

// downloadFile 负责下载文件并保存到本地
func downloadFile(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 使用Downloader来显示进度
	downloader := &Downloader{
		Reader: resp.Body,
		Total:  resp.ContentLength,
	}

	progressWriter := &ProgressWriter{
		FilePath: filePath,
		Total:    downloader.Total,
	}
	// 使用MultiWriter来同时写入文件和控制台输出进度
	multiWriter := io.MultiWriter(file, progressWriter)

	_, err = io.Copy(multiWriter, downloader)
	if err != nil {
		return err
	}

	return nil
}

// ProgressWriter 用于在控制台显示下载进度
type ProgressWriter struct {
	FilePath string
	Total    int64
	Current  int64
}

// Write 实现了io.Writer接口，用于输出进度信息
func (w *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = len(p), nil
	w.Current += int64(n)
	fmt.Printf("\r%s 下载进度：%.2f%%", w.FilePath, float64(w.Current*100)/float64(w.Total))
	if w.Current == w.Total {
		fmt.Println("    文件下载完成")
	}
	return
}

// extractFilenameFromURL 从URL中提取文件名
func extractFilenameFromURL(addr string) (filename string, err error) {
	u, err := url.Parse(addr)
	if err != nil {
		return "", err
	}
	filename = path.Base(u.Path)
	fmt.Println(addr, "提取文件名:", filename)
	return filename, nil
}
