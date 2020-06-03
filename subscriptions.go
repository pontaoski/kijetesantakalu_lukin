package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Chat struct {
	ID   int64 `gorm:"primary_key,auto_increment:false"`
	Data string
}

type PostedCommit struct {
	CommitHash           string `gorm:"primary_key"`
	ProjectWithNamespace string `gorm:"primary_key"`
}

var db *gorm.DB

func init() {
	var err error
	db, err = gorm.Open("sqlite3", "settings.db")
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&Chat{})
	db.AutoMigrate(&PostedCommit{})
}

func TrackCommit(hash, project string) bool {
	var commit PostedCommit
	db.First(&commit, "commit_hash = ?", hash, "project_with_namespace = ?", project)
	if (commit == PostedCommit{}) {
		db.Create(PostedCommit{hash, project})
		return true
	}
	return false
}

func GetChats() (ret []Chat) {
	db.Find(&ret)
	return
}

func GetChat(ID int64) (ret Chat) {
	db.First(&ret, "id = ?", ID)
	if (ret == Chat{}) {
		db.Create(Chat{ID, ""})
		ret = Chat{ID, ""}
	}
	return
}

func (c *Chat) UpdateData(data string) {
	db.Model(*c).Update("data", data)
	c.Data = data
}
