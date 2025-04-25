package unit

import "fmt"

type DataType int

const (
	Megabyte DataType = iota
	Percentage
)

func DisplayType(value float64, dataType DataType) string {
	switch dataType {
	case Megabyte:
		if value/1024 >= 10000 {
			return fmt.Sprintf("%.1f TB", value/1024/1024)
		} else if value >= 10000 {
			return fmt.Sprintf("%.1f GB", value/1024)
		} else {
			return fmt.Sprintf("%.1f MB", value)
		}
	case Percentage:
		return fmt.Sprintf("%.1f%%", value)
	}
	return ""
}
