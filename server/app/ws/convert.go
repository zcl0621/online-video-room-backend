package ws

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/openatx/go-imageio"
	"io/ioutil"
	"log"
	"online-video-room-backend/utils"
	"os"
	"strings"
)

func convert(filePath string, key string) string {
	os.MkdirAll(fmt.Sprintf("/tmp/%s", key), os.ModePerm)
	// 读取包含 base64 编码图片的文本文件
	file, _ := os.Open(filePath)
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var counter int
	var filenames []string
	// 解码每行的 base64 图片并将它们保存为文件
	for scanner.Scan() {
		txt := scanner.Text()
		if txt != "" {
			counter++
			base64png, _ := utils.UnzipString(txt)
			filename := fmt.Sprintf("/tmp/%s/img_%d.jpg", key, counter)
			base64stringToPng(base64png, filename)
			filenames = append(filenames, filename)
		}

	}
	videoFile := fmt.Sprintf("/tmp/%s.mp4", key)
	mp4 := imageio.NewVideo(videoFile, &imageio.Options{FPS: 24})

	for _, fileName := range filenames {
		fmt.Println(fileName)

		err := mp4.WriteImageFile(fileName)
		if err != nil {
			log.Printf(err.Error())
		}
	}
	mp4.Close()
	os.RemoveAll(fmt.Sprintf("/tmp/%s", key))
	os.Remove(filePath)
	return videoFile
}

func base64stringToPng(base64png string, filePath string) {
	// 解码 Base64 字符串
	imageData := strings.Split(base64png, ",")[1]
	decodedData, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		fmt.Println("解码失败:", err)
		return
	}

	// 将解码后的数据写入文件
	err = ioutil.WriteFile(filePath, decodedData, 0644)
	if err != nil {
		fmt.Println("写入文件失败:", err)
		return
	}
}
