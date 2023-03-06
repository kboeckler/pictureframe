package picture

import (
	"github.com/kboeckler/pictureframe/client"
	"github.com/kboeckler/pictureframe/config"
	"github.com/kboeckler/pictureframe/control"
	log "github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Provider interface {
	WritePicture(out io.Writer) error
}

func CreatePictureProvider(config *config.WebDavConfig, webDavClient *client.WebdavClient, control *control.Control) Provider {
	impl := &providerImpl{}
	impl.config = config
	impl.webDavClient = webDavClient
	impl.control = control
	impl.init()
	return impl
}

const MegabyteInBytes = 1024 * 1024

type providerImpl struct {
	pictures       []string
	currentPointer int
	config         *config.WebDavConfig
	webDavClient   *client.WebdavClient
	control        *control.Control
}

func (p *providerImpl) init() {
	go p.loadPictures()
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for {
			_ = <-ticker.C
			if !p.control.GetHibernate() {
				p.loadPictures()
			}
		}
	}()
}

func (p *providerImpl) loadPictures() {
	log.Println("Loading Pictures")

	foldersExcluded := make(map[string]bool)
	for _, excludedFolder := range p.config.ExcludeFolders {
		foldersExcluded[excludedFolder] = true
	}

	directories := make([]string, 0)
	for _, folder := range p.config.Folders {
		_, isExcluded := foldersExcluded[folder]
		if !isExcluded {
			directories = append(directories, "/"+folder)
		}
	}
	for {
		if len(directories) == 0 {
			break
		}
		dir := directories[0]
		directories = directories[1:]
		result, err := p.webDavClient.ReadDir(dir)
		if err != nil {
			log.Errorf("Error reading directory %s: %s", dir, err)
			continue
		}
		for _, f := range result {
			fullFilename := dir + "/" + f.Name()
			if f.IsDir() {
				_, isExcluded := foldersExcluded[fullFilename[1:]]
				if !isExcluded {
					directories = append(directories, fullFilename)
				}
			} else {
				fileSplit := strings.Split(f.Name(), ".")
				if len(fileSplit) > 1 {
					re, _ := regexp.Compile("png|jpg|jpeg|gif")
					ending := strings.ToLower(fileSplit[len(fileSplit)-1])
					fileSizeMb := float64(f.Size()) / float64(MegabyteInBytes)
					if re.MatchString(ending) && fileSizeMb <= p.config.MaxFilesizeMb {
						p.pictures = append(p.pictures, fullFilename)
					}
				}
			}
		}
	}
	sort.Sort(pictureList{p.pictures})
}

func (p *providerImpl) WritePicture(out io.Writer) error {
	if len(p.pictures) == 0 {
		log.Infof("No Pictures found\n")
		return nil
	}
	if len(p.pictures) <= p.currentPointer {
		p.currentPointer = 0
		sort.Sort(pictureList{p.pictures})
	}
	file := p.pictures[p.currentPointer]
	p.currentPointer++
	resultStream, err := p.webDavClient.ReadStream(file)
	if err != nil {
		log.Errorf("Error reading image %s: %v", file, err)
		return err
	}
	defer func(resultStream io.ReadCloser) {
		err := resultStream.Close()
		if err != nil {
			log.Warnf("Error closing file read stream from webdav: %v", err)
		}
	}(resultStream)
	_, err = io.Copy(out, resultStream)
	if err != nil {
		log.Errorf("Error writing image response %s: %v", file, err)
		return err
	}
	return nil
}

type pictureList struct {
	pictures []string
}

func (p pictureList) Len() int {
	return len(p.pictures)
}

func (p pictureList) Less(i, j int) bool {
	return rand.Intn(1) == 0
}

func (p pictureList) Swap(i, j int) {
	temp := p.pictures[i]
	p.pictures[i] = p.pictures[j]
	p.pictures[j] = temp
}
