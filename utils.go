package utils

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Create file name from ci_session cookie, timestamp and path
func CreateFileName(r *http.Request, sessionCookie string) (string, error) {
	ci_session, err := GetCookie(r, sessionCookie)
	if err != nil {
		return "", err
	}
	return MD5(r.URL.Path + ci_session + r.URL.Path + GetTimestamp()), nil
}

// Create file with md5 hash to sinalize that file was downloaded
func CreateFileFlagWithHash(filename string) error {
	tempFolder := replaceBackslashWithSlash(getTempFolderPath())
	err := os.WriteFile(concatSlash(tempFolder)+filename+"flag", []byte("1"), 0644)
	if err != nil {
		return err
	}
	return nil
}

// check if file flag exists
func CheckIfFileFlagExists(filename string) bool {
	tempFolder := replaceBackslashWithSlash(getTempFolderPath())
	if _, err := os.Stat(concatSlash(tempFolder) + filename + "flag"); err == nil {
		return true
	}
	return false
}

// Get filepath

func GetFilePath(filename string) string {
	tempFolder := replaceBackslashWithSlash(getTempFolderPath())
	return concatSlash(tempFolder) + filename
}

// Get filename from cookie and remove special caracteres and file acess bypass
func GetFilenameFromCookie(r *http.Request) (string, error) {
	filename, err := GetCookie(r, "filename")
	if err != nil {
		return "", err
	}
	return RemoveSpecialCharacters(filename), nil
}

// Delete file flag
func DeleteFileFlag(filename string) error {
	tempFolder := replaceBackslashWithSlash(getTempFolderPath())
	err := os.Remove(concatSlash(tempFolder) + filename + "flag")
	if err != nil {
		return err
	}
	return nil
}

// Get timestamp
func GetTimestamp() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

// // Get ci_session
// func getCISessionCookie(r *http.Request) (string, error) {
// 	return GetCookie(r, "ci_session")
// }

// Get cookie
func GetCookie(r *http.Request, cookieName string) (string, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// MD5 hash from string
func MD5(text string) string {
	hash := md5.New()
	hash.Write([]byte(text))
	return RemoveSpecialCharacters(Base64Encode(hash.Sum(nil)))
}

// remove special characters from string
func RemoveSpecialCharacters(text string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	processedString := reg.ReplaceAllString(text, "")
	return processedString
}

// base64 encode from []byte
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// save response body to file
func SaveResponseBodyToFile(filename string, recorder *httptest.ResponseRecorder, Log *zap.Logger) {
	body := recorder.Body.Bytes()
	tempFolder := replaceBackslashWithSlash(getTempFolderPath())
	err := os.WriteFile(concatSlash(tempFolder)+filename, body, 0644)
	if err != nil {
		Log.Error("Error writing file", zap.Error(err))
	}
}

// get temporary folder path windows and linux
func getTempFolderPath() string {
	return os.TempDir()
}

// replace \ with / in path
func replaceBackslashWithSlash(path string) string {
	return strings.Replace(path, "\\", "/", -1)
}

// if string last position not is / concat /
func concatSlash(path string) string {
	if path[len(path)-1:] != "/" {
		return path + "/"
	}
	return path
}

// convert string to int and return error if not possible
func ConvertStringToInt(text string) (int, error) {
	return strconv.Atoi(text)
}

// convert int to string
func ConvertIntToString(number int) string {
	return strconv.Itoa(number)
}

// delete file
func DeleteFile(filename string) error {
	return os.Remove(filename)
}

// Create configuration from environment variables
func CreateConfigFromEnv(log *zap.Logger) *Configuration {
	return &Configuration{
		LogLevel:          verifyLogLevel(os.Getenv("LOG_LEVEL")),
		VisibleHeaders:    verifyVisibleHeaders(os.Getenv("VISIBLE_HEADERS")),
		HeaderToReplicate: verifyHeaderToReplicate(os.Getenv("HEADER_TO_REPLICATE")),
		BodyMaxLen:        verifyBodyMaxLen(os.Getenv("LOG_BODY_MAX_SIZE"), log),
		Proxy:             ConvertJSONToRouter(os.Getenv("PROXY_CONFIGURATION"), log),
		Listener:          verifyListener(os.Getenv("LISTEN_PORT"), os.Getenv("LISTEN_HOST"), log),
	}
}

// Append HEADER_TO_REPLICATE to Headers type
func (h Headers) append(header string) Headers {
	return append(h, header)
}

// Verify if HEADER_TO_REPLICATE is valid and return Headers type result from HEADER_TO_REPLICATE splitted by ","
// If not, not set.
func verifyHeaderToReplicate(headerToReplicate string) Headers {
	if headerToReplicate == "" {
		return Headers{}
	}
	var headers Headers
	//Check if each one is valid
	for _, header := range strings.Split(headerToReplicate, ",") {
		if header != "" {
			headers = headers.append(header)
		}
	}
	return headers
}

// Verify LISTEN_PORT and LISTEN_HOST and return Listener struct
// If not, set to "8080" and "localhost"
func verifyListener(listenPort string, listenHost string, log *zap.Logger) Listener {
	if listenPort == "" {
		listenPort = "8080"
	}
	if listenHost == "" {
		listenHost = "localhost"
	}
	listenPortInt, err := ConvertStringToInt(listenPort)
	if err != nil {
		log.Warn("Invalid LISTEN_PORT")
		listenPortInt = 8080
	}
	return Listener{
		Address: listenHost,
		Port:    ConvertIntToString(listenPortInt),
	}
}

// Verify if LOG_BODY_MAX_SIZE is valid
// If not, set to 255
func verifyBodyMaxLen(bodyMaxLen string, log *zap.Logger) int {
	if bodyMaxLen == "" {
		log.Warn("Invalid LOG_BODY_MAX_SIZE")
		return 255
	}
	if bodyManLenInt, err := ConvertStringToInt(bodyMaxLen); err != nil {
		log.Warn("Invalid LOG_BODY_MAX_SIZE")
		return 255
	} else {
		return bodyManLenInt
	}
}

// Verify if VisibleHeaders is valid
// If not, set to "host,x-request-id,x-real-ip,content-length,user-agent,accept-encoding,content-type,custom-app-headers"
func verifyVisibleHeaders(visibleHeaders string) string {
	if visibleHeaders == "" {
		return "host,x-request-id,x-real-ip,content-length,user-agent,accept-encoding,content-type,custom-app-headers"
	}
	return visibleHeaders
}

// Verify if LOG_LEVEL is valid
// If not, set to "info"
func verifyLogLevel(logLevel string) string {
	switch strings.ToLower(logLevel) {
	case "debug":
		return "debug"
	case "info":
		return "info"
	case "warn":
		return "warn"
	case "error":
		return "error"
	default:
		return "info"
	}
}

// Convert JSON to list of router struct
// [{"host": "localhost", "routes": ["/teste.txt"], "mode": 0}]
// [{"host": "localhost", "routes": ["/teste.txt"], "mode": 1}]
func ConvertJSONToRouter(jsonData string, log *zap.Logger) []Router {
	var routers []Router
	err := json.Unmarshal([]byte(jsonData), &routers)
	if err != nil {
		log.Fatal("Error parsing JSON", zap.Error(err))
	}
	return routers
}
