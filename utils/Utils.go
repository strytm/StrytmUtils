package utils

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/dchest/uniuri"
	"github.com/go-resty/resty/v2"
	"github.com/tcolgate/mp3"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	_                     = iota
	KB                int = 1 << (10 * iota)
	MB
	GB
	RandomChar = "abcdefghijklmnopqrstuvwxyz0123456789_"
)

type Sizer interface {
	Size() int64
}

type AudioMeta struct {
	Name    string
	Bitrate int
}

type FileBytesMeta struct {
	Reader *bytes.Reader
	Name string
}

func GetAllFormRequestValue(r *http.Request) map[string]interface{} {
	clearMapData := make(map[string]interface{})


	// chon r.Form tamame maghadir ro to array mirikht on maghidir ro az array dar avrodam
	for i, value := range r.Form {

		clearMapData[i] = value[0]

	}

	return clearMapData
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}



func WriteAllPostImageFromRequest(r *http.Request, keyFileValue string, path string , maxWith int , maxHeight int , maxSize int) (chan string, error) {

	err := r.ParseMultipartForm(32 << 20) //32 MB

	if err != nil{
		return nil, err
	}

	allFile := r.MultipartForm.File[keyFileValue]

	fileCount := len(allFile)

	if fileCount > 0 {

		if err := FolderMaker(path); err != nil {
			fmt.Println(err.Error())
			return nil, err
		}

		fileNameChan := make(chan string, fileCount)

		for _, value := range allFile {

			if fileName, errImageWriter := ImageWriterByFileHeader(value , path , maxWith , maxHeight , maxSize ); errImageWriter != nil{
				return nil, errImageWriter
			}else
			{
				fileNameChan <- fileName
			}

		}

		return fileNameChan, nil

	}


	return nil, nil

}


func ImageWriterByFileHeader(fileHeader *multipart.FileHeader, path string , maxWith int , maxHeight int , maxSize int) (string, error) {

	file, err := fileHeader.Open()

	if err != nil{
		return "", err
	}

	buffer := make([]byte, fileHeader.Size)

	for {

		value, err := file.Read(buffer)

		if err != nil && err != io.EOF {
			return "", err
		}
		if value == 0 {
			break
		}
	}

	if err := file.Close(); err != nil {
		return "", err
	}

	conf, format, err := image.DecodeConfig(bytes.NewReader(buffer))
	if err != nil {
		return "", err
	}

	if format != "jpeg" && format != "png" && format != "jpg" {
		return "", errors.New(" format haye jpeg , png , jpg pazirofte mishavad")
	}

	if conf.Height > maxHeight || conf.Width > maxWith {

		return "", errors.New(fmt.Sprintf("tasvir bayad kochaktar az %d x %d pixel bashand", maxWith , maxHeight))
	}

	if len(buffer) >= maxSize*MB {
		return "", errors.New(fmt.Sprintf("tasvir bayad kochaktar az %d hajm dashte bashand", MB))
	}

	fileName := uniuri.NewLenChars(10, []byte(RandomChar)) + fmt.Sprint(time.Now().Unix())

	if errIo := ioutil.WriteFile(path+fileName+"."+format, buffer, 0700); errIo != nil {
		return "", errIo
	}

	return fileName + "." + format, nil
}


func WriteAllPostAudioFromRequest(r *http.Request, keyFileValue string, path string , maxSize int) (chan AudioMeta, error) {

	err := r.ParseMultipartForm(32 << 20) //32 MB

	if err != nil {
		return nil, err
	}

	allFile := r.MultipartForm.File[keyFileValue]

	fileCount := len(allFile)

	if fileCount > 0 {


		fileNameChan := make(chan AudioMeta, fileCount)

		for _, value := range allFile {


			if fileName, errAudioWriter := AudioWriterByFileHeader(value , path , maxSize ); errAudioWriter != nil{
				return nil, errAudioWriter
			}else
			{
				fileNameChan <- fileName
			}

		}

		return fileNameChan, nil

	}

	return nil, nil
}


func WriteAllPostVideoFromRequest(r *http.Request, keyFileValue string, path string , maxSize int) (chan string, error) {

	err := r.ParseMultipartForm(32 << 20) //32 MB

	if err != nil {
		return nil, err
	}

	allFile := r.MultipartForm.File[keyFileValue]

	fileCount := len(allFile)

	if fileCount > 0 {

		if err := FolderMaker(path); err != nil {
			fmt.Println(err.Error())
			return nil, err
		}

		fileNameChan := make(chan string, fileCount)

		for _, value := range allFile {


			if fileName, errAudioWriter := VideoWriterByFileHeader(value , path , maxSize ); errAudioWriter != nil{
				return nil, errAudioWriter
			}else
			{
				fileNameChan <- fileName
			}

		}

		return fileNameChan, nil

	}

	return nil, nil
}


func VideoWriterByFileHeader(fileHeader *multipart.FileHeader,  path string, maxSize int)(string, error) {

	format := fileHeader.Filename[len(fileHeader.Filename)-3:]

	if format != "mp4" {
		return "", errors.New(" format mp4 pazirofte mishavad - format ersali : "+format )
	}

	file, err := fileHeader.Open()

	if err != nil{
		return "", err
	}

	buffer := make([]byte, fileHeader.Size)

	for {

		value, err := file.Read(buffer)

		if err != nil && err != io.EOF {
			return "", err
		}
		if value == 0 {
			break
		}
	}

	if err := file.Close(); err != nil {
		return "", err
	}


	if len(buffer) >= maxSize*MB {
		return "", errors.New(fmt.Sprintf("file bayad kochaktar az %d MB hajm dashte bashand", maxSize))
	}

	fileName := uniuri.NewLenChars(10, []byte(RandomChar)) + fmt.Sprint(time.Now().Unix())


	if errIo := ioutil.WriteFile(path+fileName+"."+format, buffer, 0700); errIo != nil {
		return "", errIo
	}

	return fileName + "." + format, nil
}


func AudioWriterByFileHeader(fileHeader *multipart.FileHeader,  path string, maxSize int)(AudioMeta, error) {

	audioMeta := AudioMeta{}

	format := fileHeader.Filename[len(fileHeader.Filename)-3:]

	if format != "mp3" && format != "wav" && format != "aac" {
		return audioMeta, errors.New(" format haye mp3 , wav , aac pazirofte mishavad - formate ersali : "+format)
	}

	file, err := fileHeader.Open()

	if err != nil{
		return audioMeta, err
	}

	buffer := make([]byte, fileHeader.Size)

	for {

		value, err := file.Read(buffer)

		if err != nil && err != io.EOF {
			return audioMeta, err
		}
		if value == 0 {
			break
		}
	}

	if err := file.Close(); err != nil {
		return audioMeta, err
	}


	ioReader := bytes.NewReader(buffer)

	mp3BitrateDecoder := mp3.NewDecoder(ioReader)
	var mp3Frame mp3.Frame
	skipped := 0
	if err := mp3BitrateDecoder.Decode(&mp3Frame, &skipped); err != nil {
		return audioMeta, err
	}

	mp3Bitrate := mp3Frame.Header().BitRate() / 1000

	path = path + fmt.Sprint(mp3Bitrate) + "/"
	if err := FolderMaker(path); err != nil {
		fmt.Println(err.Error())
		return audioMeta, err
	}


	if len(buffer) >= maxSize*MB {
		return audioMeta, errors.New(fmt.Sprintf("file bayad kochaktar az %d MB hajm dashte bashand", maxSize))
	}

	fileName := uniuri.NewLenChars(10, []byte(RandomChar)) + fmt.Sprint(time.Now().Unix())


	if errIo := ioutil.WriteFile(path+fileName+"."+format, buffer, 0700); errIo != nil {
		return audioMeta, errIo
	}

	audioMeta.Name = fileName + "." + format
	audioMeta.Bitrate = int(mp3Bitrate)

	return audioMeta, nil
}

func FolderMaker(path string) error {

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if errMkDir := os.MkdirAll(path, 0700); err != nil {
				return errMkDir
			}
		}
	} else {
		return err
	}

	return nil
}

func ExteraxtTokenFromHeader(key string, r *http.Request) (string, error) {

	authorizationValue := r.Header.Get("Authorization")

	if len(authorizationValue) == 0{

		return "" , errors.New("token bayad be sorate Bearer Token ersal shavad")

	}else
	{
		bearerTokenSlice := strings.Split(authorizationValue , " ")
		if bearerTokenSlice[0] != "Bearer"{
			return "" , errors.New("kalameye kelidye Bearer ersal nashode ast")
		}
		return bearerTokenSlice[1], nil

	}

}

func StringToUint(value string) uint64 {
	uintValue, _ := strconv.ParseUint(value, 10, 32)

	return uintValue
}



func PostAllFileToThisURL(r *http.Request , fileKey string,formDataMap map[string]string , url string) (string , error) {

	err := r.ParseMultipartForm(32 << 20) //32 MB

	if err != nil{
		return  "" ,err
	}

	var fileByteData []FileBytesMeta

	for _, fileHeader := range r.MultipartForm.File[fileKey] {

		file, err := fileHeader.Open()

		if err != nil {
			return "" ,err
		}

		byteArray, err := ioutil.ReadAll(file)
		if err != nil {
			return "" ,err
		}

		fileBytesMeta := FileBytesMeta{}
		fileBytesMeta.Name = fileHeader.Filename
		fileBytesMeta.Reader = bytes.NewReader(byteArray)

		fileByteData = append(fileByteData, fileBytesMeta)
	}

	if len(fileByteData) != 0 {

		return postBytesToThisURL(fileByteData , fileKey , formDataMap , url)
	}


	return "" ,nil
}

func postBytesToThisURL(fileByteData []FileBytesMeta , key string , formDataMap map[string]string , url string) (string ,error) {
	restyClient := resty.New().R()

	for _, value := range fileByteData {
		restyClient.SetFileReader(key , value.Name , value.Reader)
	}

	response, err := restyClient.SetFormData(formDataMap).Post(url)
	if err != nil {
		return "" ,err
	}
	defer func() {restyClient = nil}()

	if response.StatusCode() != http.StatusOK {
		return "",errors.New("error status "+fmt.Sprint(response.String()))
	}else
	{
		return response.String() , nil
	}

}

func StringSplitterToUint(value string, splitter string) ([]uint, error) {
	var uintValueArray []uint

	if len(value) > 0 {
		strArray := strings.Split(value, splitter)
		for _, s := range strArray {
			if convertedValue, err := strconv.ParseUint(s, 10, 32); err != nil {
				return nil, errors.New("meghdare vorody adad sahih nemibashad")
			} else {
				uintValueArray = append(uintValueArray, uint(convertedValue))
			}
		}
	}

	return uintValueArray, nil
}

func DifferentIds(list interface{}, ids []uint) error {
	founded := false

	valueOfData := reflect.ValueOf(list)

	if valueOfData.Len() != len(ids) {
		for i := 0; i < len(ids); i++ {
			founded = false
			for j := 0; j < valueOfData.Len(); j++ {
				if ids[i] == uint(reflect.Indirect(valueOfData.Index(j)).FieldByName("ID").Uint()) {
					founded = true
					break
				}
			}

			if !founded {

				return errors.New("id e "+fmt.Sprint(ids[i])+" vojod nadard")
			}
		}
	}
	return nil
}

func ServiceLocator(url string) (ServiceLocatorModel , error){
	restyClient := resty.New().R()

	serviceLocatorModel := ServiceLocatorModel{}

	if response, err := restyClient.SetResult(&serviceLocatorModel).Get(url); err != nil {
		return serviceLocatorModel , err
	} else
	{
		if response.StatusCode() != http.StatusOK{
			return serviceLocatorModel , errors.New(response.String())
		}else
		{
			return serviceLocatorModel , nil
		}

	}

}
