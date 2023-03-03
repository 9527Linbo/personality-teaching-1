package mysql

import (
	"fmt"
	"personality-teaching/src/model"

	"gorm.io/gorm"
)

const (
	classValid   int8 = 1 //合法记录
	classInvalid int8 = 0 //不合法记录
)

type classFunc interface {
	InsertClass(teacherID string, c model.Class) error
	UpdateClass(m model.Class) error
	DeleteClass(teacherID string, classID string) error
	QueryClass(classID string) (model.Class, error)
	QueryClassList(teacherID string, req model.ClassListReq) ([]model.Class, int, error)
	CheckTeacherClass(teacherID string, classID string) (bool, error)
	CheckClassName(name string) (bool, error)
}

type ClassMySQL struct{}

var _ classFunc = &ClassMySQL{}

func NewClassMysql() *ClassMySQL {
	return &ClassMySQL{}
}

func (c *ClassMySQL) InsertClass(teacherID string, m model.Class) error {
	return Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("insert into `t_class`(`class_id`,`name`,`college`,`major`) values (?,?,?,?)",
			m.ClassID, m.Name, m.College, m.Major).Error; err != nil {
			return err
		}
		if err := tx.Exec("insert into `t_teacher_class`(`class_id`,`teacher_id`,`is_valid`) values (?,?,?)",
			m.ClassID, teacherID, classValid).Error; err != nil {
			return err
		}
		return nil
	})

}

func (c *ClassMySQL) UpdateClass(m model.Class) error {
	return Db.Exec("update `t_class` set `name` = ?,`college` = ?,`major` = ? where class_id = ?",
		m.Name, m.College, m.Major, m.ClassID).Error
}

func (c *ClassMySQL) DeleteClass(teacherID string, classID string) error {
	return Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("delete from `t_class` where `class_id` = ?", classID).Error; err != nil {
			return err
		}
		if err := tx.Exec("update `t_teacher_class` set `is_valid` = ? where `teacher_id` = ? and `class_id` = ?",
			classInvalid, teacherID, classID).Error; err != nil {
			return err
		}
		return nil
	})
}

func (c *ClassMySQL) QueryClass(classID string) (model.Class, error) {
	var m model.Class
	if err := Db.Raw("select `class_id`,`name`,`college`,`major` from `t_class` where class_id = ?", classID).Scan(&m).Error; err != nil {
		return model.Class{}, err
	}
	return m, nil
}

func (c *ClassMySQL) QueryClassList(teacherID string, req model.ClassListReq) ([]model.Class, int, error) {
	var classes []model.Class
	var total int
	offset := (req.PageNum - 1) * req.PageSize
	count := req.PageSize
	err := Db.Raw("select `t_class`.`class_id`,`name`,`college`,`major` from `t_class` inner join `t_teacher_class` "+
		"on `t_class`.class_id = `t_teacher_class`.class_id "+
		"where teacher_id = ? and `is_valid` = ? limit ?,?", teacherID, classValid, offset, count).Scan(&classes).Error
	if err != nil {
		return []model.Class{}, 0, err
	}
	err = Db.Raw("select count(`id`) from `t_teacher_class` where `teacher_id` = ? and `is_valid` = ?", teacherID, classValid).Scan(&total).Error
	if err != nil {
		return []model.Class{}, 0, err
	}
	return classes, total, nil
}

// CheckTeacherClass 有此数据返回true
func (c *ClassMySQL) CheckTeacherClass(teacherID string, classID string) (bool, error) {
	id := ""
	err := Db.Raw("select `id` from `t_teacher_class` where `teacher_id` = ? and `class_id` = ? and `is_valid` = ?", teacherID, classID, classValid).Scan(&id).Error
	if err != nil {
		return false, err
	}
	return !(id == ""), nil
}

// CheckClassName 有此数据返回true
func (c *ClassMySQL) CheckClassName(name string) (bool, error) {
	var classID string
	name = fmt.Sprintf("%%%s%%", name)
	if err := Db.Raw("select `class_id` from `t_class` where `name` like ?;", name).Scan(&classID).Error; err != nil {
		return false, err
	} else if classID == "" {
		return false, nil
	}
	return true, nil
}
