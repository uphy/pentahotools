package client

import (
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/resty.v0"
)

type DownloadHelper struct {
	overwrite    bool
	destination  string
	tmpFile      string
	FilenameFunc func(*resty.Response) string
}

func NewDownloadHelper(destination string, overwrite bool) *DownloadHelper {
	return &DownloadHelper{
		destination: destination,
		overwrite:   overwrite,
	}
}

func (h *DownloadHelper) GetTemporaryFilePath() string {
	return h.tmpFile
}

func (h *DownloadHelper) existsAndIsFile(file string) bool {
	s, err := os.Stat(file)
	return err == nil && s.IsDir() == false
}

func (h *DownloadHelper) PrepareTemporaryFile() error {
	if h.existsAndIsFile(h.destination) && !h.overwrite {
		return errors.New("destination file already exist")
	}
	tmp, err := ioutil.TempFile("", "download")
	if err != nil {
		return errors.Wrap(err, "create temp file failed")
	}
	defer tmp.Close()
	h.tmpFile = tmp.Name()
	return nil
}

func (h *DownloadHelper) findFilenameFromContentDisposition(contentDisposition string) string {
	pattern, _ := regexp.Compile(`attachment; filename\*=UTF-8''(.*)`)
	encodedFilename := pattern.FindStringSubmatch(contentDisposition)
	if len(encodedFilename) == 2 {
		filename, _ := url.PathUnescape(encodedFilename[1])
		return filename
	}
	pattern, _ = regexp.Compile(`attachment; filename="(.*)"`)
	filename := pattern.FindStringSubmatch(contentDisposition)
	if len(filename) == 2 {
		return filename[1]
	}
	return "downloadedfile.bin"
}

func (h *DownloadHelper) MoveTemporaryFileToDestination(resp *resty.Response) (string, error) {
	fixedDestination := h.destination
	stat, err := os.Stat(fixedDestination)
	if fixedDestination == "" || (os.IsExist(err) && stat.IsDir()) {
		var filename string
		if h.FilenameFunc == nil {
			contentDisposition := resp.Header().Get("Content-Disposition")
			filename = h.findFilenameFromContentDisposition(contentDisposition)
		} else {
			filename = h.FilenameFunc(resp)
		}
		if strings.HasSuffix(fixedDestination, "/") {
			fixedDestination = fixedDestination + filename
		} else {
			if len(fixedDestination) == 0 {
				fixedDestination = "./" + filename
			} else {
				fixedDestination = fixedDestination + "/" + filename
			}
		}
	}
	if h.existsAndIsFile(fixedDestination) && !h.overwrite {
		return "", errors.New("destination file already exist")
	}
	err = h.move(h.GetTemporaryFilePath(), fixedDestination)
	return fixedDestination, nil
}

func (h *DownloadHelper) move(from string, to string) error {
	reader, err := os.Open(from)
	if err != nil {
		return err
	}
	defer reader.Close()
	writer, err := os.Open(to)
	if err != nil {
		return err
	}
	defer writer.Close()
	_, err = io.Copy(writer, reader)
	if err != nil {
		return err
	}
	err = os.Remove(from)
	return err
}

func (h *DownloadHelper) Clean() {
	// do nothing
}
