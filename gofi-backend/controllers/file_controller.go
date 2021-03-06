package controllers

import (
	"github.com/kataras/iris"
	"github.com/sirupsen/logrus"
	"gofi/context"
	"gofi/env"
	"gofi/i18n"
	"gofi/models"
	"gofi/util"
	"io/ioutil"
	"path/filepath"
	"strings"
)

//ListFiles 返回给定路径文件夹的一级子节点文件
func ListFiles(ctx iris.Context) {
	// 需要列出文件的文件夹地址相对路径
	relativePath := ctx.URLParamDefault("path", "")

	storagePath := context.Get().GetStorageDir()

	logrus.Printf("root path is %v \n", storagePath)

	path := filepath.Join(storagePath, relativePath)

	// 确保该路径只是文件仓库的子路径
	if !strings.Contains(path, storagePath) {
		ctx.JSON(ResponseFailWithMessage(i18n.Translate(i18n.OperationNotAllowedInPreviewMode)))
		return
	}

	if !util.FileExist(path) {
		ctx.JSON(ResponseFailWithMessage(i18n.Translate(i18n.DirIsNotExist, path)))
		return
	}

	if !util.IsDirectory(path) {
		ctx.JSON(ResponseFailWithMessage(i18n.Translate(i18n.IsNotDir, path)))
		return
	}

	// 读取该文件夹下所有文件
	files, err := ioutil.ReadDir(path)

	// 读取失败
	if err != nil {
		ctx.JSON(ResponseFailWithMessage(err.Error()))
		return
	}

	var filesOfDir []models.File

	// 将所有文件再次封装成客户端需要的数据格式
	for _, f := range files {

		// 当前文件是隐藏文件(以.开头)则不显示
		if util.IsHiddenFile(f.Name()) {
			continue
		}

		var size int
		if f.IsDir() {
			size = 0
		} else {
			size = int(f.Size())
		}

		// 实例化File model
		file := models.File{
			IsDirectory:  f.IsDir(),
			Name:         f.Name(),
			Size:         size,
			Extension:    strings.TrimLeft(filepath.Ext(f.Name()), "."),
			Path:         filepath.Join(relativePath, f.Name()),
			LastModified: f.ModTime().Unix(),
		}

		// 添加到切片中等待json序列化
		filesOfDir = append(filesOfDir, file)
	}

	ctx.JSON(ResponseSuccess(filesOfDir))

	return
}

//Upload 上传文件
func Upload(ctx iris.Context) {
	// 预览模式禁止上传文件
	if env.Current == env.Preview {
		ctx.JSON(ResponseFailWithMessage(i18n.Translate(i18n.CurrentIsPreviewMode)))
		return
	}

	// 需要列出文件的文件夹地址相对路径
	relativePath := ctx.URLParamDefault("path", "")

	logrus.Infof("relativePath path is %v \n", relativePath)

	storageDir := context.Get().GetStorageDir()

	logrus.Infof("root path is %v \n", storageDir)

	destDirectory := filepath.Join(storageDir, relativePath)

	logrus.Infof("destPath path is %v \n", destDirectory)

	err := ctx.Request().ParseMultipartForm(ctx.Application().ConfigurationReadOnly().GetPostMaxMemory())
	if err != nil {
		ctx.JSON(ResponseFailWithMessage(i18n.Translate(i18n.UploadFailed)))
		return
	}

	if ctx.Request().MultipartForm != nil {
		if fileHeaders := ctx.Request().MultipartForm.File; fileHeaders != nil {
			for _, files := range fileHeaders {
				for _, file := range files {

					if util.FileExist(filepath.Join(destDirectory, file.Filename)) {
						ctx.JSON(ResponseFailWithMessage(i18n.Translate(i18n.CanNotOverlayExistFile, file.Filename)))
						return
					}

					_, err := util.UploadFileTo(file, destDirectory)
					if err != nil {
						ctx.JSON(ResponseFailWithMessage(i18n.Translate(i18n.UploadFailed)))
						return
					}
				}
			}
		}
	}

	ctx.JSON(ResponseSuccess(nil))
	return
}

//Download 下载文件
func Download(ctx iris.Context) {
	ctx.ReadJSON(map[string]interface{}{"Key": "value"})
	// 需要列出文件的文件夹地址相对路径
	relativePath := ctx.URLParamDefault("path", "")

	storageDir := context.Get().GetStorageDir()

	logrus.Printf("root path is %v \n", storageDir)

	path := filepath.Join(storageDir, relativePath)

	if !util.FileExist(path) {
		ctx.JSON(ResponseFailWithMessage(i18n.Translate(i18n.FileIsNotExist, path)))
		return
	}

	if util.IsDirectory(path) {
		ctx.JSON(ResponseFailWithMessage(i18n.Translate(i18n.IsNotFile, path)))
		return
	}

	filename := filepath.Base(path)

	ctx.SendFile(path, filename)
}
