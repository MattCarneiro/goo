package main

import (
  "fmt"
  "log"
  "net/http"
  "os"
  "regexp"

  "github.com/gin-gonic/gin"
  "google.golang.org/api/drive/v3"
  "google.golang.org/api/googleapi"
  "golang.org/x/oauth2/google"
)

var (
  mimeTypes = map[string]string{
    "pdf":   "application/pdf",
    "image": "image/",
    "video": "video/",
  }
  apiKey string
)

func main() {
  apiKey = os.Getenv("GOOGLE_DRIVE_API_KEY")
  if apiKey == "" {
    log.Fatal("GOOGLE_DRIVE_API_KEY environment variable not set")
  }

  r := gin.Default()
  r.POST("/check-downloadable", checkDownloadableHandler)
  
  port := os.Getenv("PORT")
  if port == "" {
    port = "3000"
  }
  
  log.Printf("Server running at http://localhost:%s", port)
  log.Fatal(r.Run(":" + port))
}

func checkDownloadableHandler(c *gin.Context) {
  var request struct {
    Link string `json:"link"`
    Type string `json:"type"`
  }
  
  if err := c.ShouldBindJSON(&request); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameters"})
    return
  }

  if _, ok := mimeTypes[request.Type]; !ok {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type"})
    return
  }

  id := extractIdFromLink(request.Link)
  if id == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid link format"})
    return
  }

  isFolder := regexp.MustCompile(`/folders/`).MatchString(request.Link)
  var downloadable bool
  var err error

  if isFolder {
    downloadable, err = checkFolder(id, request.Type)
  } else {
    downloadable, err = isDownloadable(id, request.Type)
  }

  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }

  result := "no"
  if downloadable {
    result = "yes"
  }

  c.JSON(http.StatusOK, gin.H{"result": result})
}

func isDownloadable(fileId, fileType string) (bool, error) {
  srv, err := drive.NewService(c)
  if err != nil {
    return false, fmt.Errorf("unable to create Drive service: %v", err)
  }

  file, err := srv.Files.Get(fileId).Fields("mimeType").Do()
  if err != nil {
    return false, fmt.Errorf("unable to get file: %v", err)
  }

  switch fileType {
  case "pdf":
    return file.MimeType == mimeTypes["pdf"], nil
  case "image":
    return strings.HasPrefix(file.MimeType, mimeTypes["image"]), nil
  case "video":
    return strings.HasPrefix(file.MimeType, mimeTypes["video"]), nil
  default:
    return false, nil
  }
}

func checkFolder(folderId, fileType string) (bool, error) {
  srv, err := drive.NewService(c)
  if err != nil {
    return false, fmt.Errorf("unable to create Drive service: %v", err)
  }

  files, err := srv.Files.List().
    Q(fmt.Sprintf("'%s' in parents", folderId)).
    Fields("files(id, mimeType)").
    Do()
  if err != nil {
    return false, fmt.Errorf("unable to list files: %v", err)
  }

  if len(files.Files) == 0 {
    return false, nil
  }

  for _, file := range files.Files {
    switch fileType {
    case "pdf":
      if file.MimeType == mimeTypes["pdf"] {
        return true, nil
      }
    case "image":
      if strings.HasPrefix(file.MimeType, mimeTypes["image"]) {
        return true, nil
      }
    case "video":
      if strings.HasPrefix(file.MimeType, mimeTypes["video"]) {
        return true, nil
      }
    }
  }

  return false, nil
}

func extractIdFromLink(link string) string {
  fileIdRegex := regexp.MustCompile(`/d/([a-zA-Z0-9-_]+)`)
  folderIdRegex := regexp.MustCompile(`/folders/([a-zA-Z0-9-_]+)`)

  if fileIdMatch := fileIdRegex.FindStringSubmatch(link); len(fileIdMatch) > 1 {
    return fileIdMatch[1]
  }
  if folderIdMatch := folderIdRegex.FindStringSubmatch(link); len(folderIdMatch) > 1 {
    return folderIdMatch[1]
  }
  return ""
}
