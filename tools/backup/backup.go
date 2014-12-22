package backup

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	STARTOFFSET = 1 << 6
)

type BackupMap struct {
	getdata func() ([]byte, error)
	putdata func([]byte) error
}

type backupHeaderT struct {
	Size   int
	Offset int
	Name   string
}

type BackupCronT map[string]BackupMap

var BackupCron = make(BackupCronT)

func (self *BackupCronT) AddTask(name string, getter func() ([]byte, error), putter func([]byte) error) {
	(*self)[name] = BackupMap{getter, putter}
}

func (self BackupCronT) Restore(file string) (readed int, err error) {
	fd, err := os.Open(file)
	if err != nil {
		return
	}
	defer fd.Close()
	readBuff := make([]byte, 4096)
	readed, err = fd.Read(readBuff)
	if err != nil {
		return
	}
	log.Println("[BACKUP RESTORE INFO] readed: ", readed)
	headerSize, err := binary.ReadUvarint(bytes.NewReader(readBuff[:STARTOFFSET]))
	if err != nil {
		return
	}
	log.Println("[BACKUP RESTORE INFO] HeaderSize: ", headerSize)

	header := make([]backupHeaderT, 0)
	buffer := bytes.NewBuffer(readBuff[STARTOFFSET:(STARTOFFSET + headerSize)])
	decoder := gob.NewDecoder(buffer)
	if err = decoder.Decode(&header); err != nil {
		log.Println("[BACKUP RESTORE ERROR]", err.Error())
		return
	}
	for num, _ := range header {
		if err := BackupCron[header[num].Name].putdata(readBuff[STARTOFFSET+headerSize : STARTOFFSET+headerSize+uint64(header[0].Size)]); err != nil {
			log.Println("[BACKUP RESTORE ERROR]", err.Error())
		}
	}
	return
}

func (self *BackupCronT) Start(period int, file string) error {
	go func() {
		for {
			ticker := time.NewTicker(time.Duration(period) * time.Second)
			<-ticker.C
			backupBuffer := new(bytes.Buffer)
			header := make([]backupHeaderT, 0)
			offset := 0
			for key, _ := range *self {
				data, err := (*self)[key].getdata()
				if err != nil {
					log.Println("[BACKUP ERROR] ", err.Error())
				}
				size, err := backupBuffer.Write(data)
				if err != nil {
					log.Println("[BACKUP ERROR] ", err.Error())
				}
				header = append(header, backupHeaderT{Size: size, Offset: offset, Name: key})
			}

			headerGob := new(bytes.Buffer)
			enc := gob.NewEncoder(headerGob)
			if err := enc.Encode(header); err != nil {
				log.Println("[BACKUP ERROR] encode error:", err)
			}

			headerSize := uint64(len(headerGob.Bytes()))
			preHeader := make([]byte, STARTOFFSET-1)
			binary.PutUvarint(preHeader, uint64(headerSize))
			log.Println(headerGob.Bytes())
			log.Println(header)

			res, err := writeToDump(file, preHeader, headerGob.Bytes(), backupBuffer.Bytes())
			if err != nil {
				log.Println("[BACKUP ERROR] " + err.Error())
			} else {
				log.Println("[BACKUP INFO] Backup wrote " + strconv.Itoa(res) + " bytes")
			}
		}
	}()
	return nil
}

func writeToDump(file string, preheader []byte, header []byte, data []byte) (res int, err error) {
	fd, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0660)
	writeRes := 0
	if err != nil {
		return
	}
	defer fd.Close()
	if writeRes, err = fd.Write(preheader); err != nil {
		return
	}
	res += writeRes
	if writeRes, err = fd.WriteAt(header, STARTOFFSET); err != nil {
		res += writeRes
		return
	}
	res += writeRes
	if writeRes, err = fd.WriteAt(data, (int64)(STARTOFFSET+writeRes)); err != nil {
		res += writeRes
		return
	}
	res += writeRes
	return
}
