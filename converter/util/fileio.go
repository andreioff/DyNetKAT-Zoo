package util

import "os"

const FILE_PERM = 0755

func WriteToNewFile(dir, fileName, data string) error {
	_, err := os.Stat(dir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if os.IsNotExist(err) {
		err = os.Mkdir(dir, FILE_PERM)
	}

	filePath := dir + fileName
	if dir[len(dir)-1] != '/' {
		filePath = dir + "/" + fileName
	}

	os.WriteFile(filePath, []byte(data), FILE_PERM)
	return nil
}
