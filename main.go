package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type Arguments map[string]string

type Item struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func parseItem(item string) (id string, email string, age int) {
	splitedItem := strings.Split(item, ",")
	id = getID(splitedItem)
	email = getEmail(splitedItem)
	age = getAge(splitedItem)

	return id, email, age
}

func getID(splitedItem []string) string {
	id := strings.Split(splitedItem[0], ":")[1]
	idUnquoted, _ := strconv.Unquote(id)
	return idUnquoted
}

func getEmail(splitedItem []string) string {
	email := strings.Split(splitedItem[1], ":")[1]
	emailUnquoted, _ := strconv.Unquote(email)
	return emailUnquoted
}

func getAge(splitedItem []string) int {
	age := strings.Split(splitedItem[2], ":")[1]
	intAge, _ := strconv.Atoi(age)
	return intAge
}

func readFromFile(fileName string) (items []Item, byteValue []byte) {

	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}

	byteValue, _ = ioutil.ReadAll(file)
	json.Unmarshal(byteValue, &items)
	file.Close()
	return items, byteValue
}

func writeToFile(fileName string, items []Item) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}

	encodedItems, _ := json.Marshal(items)

	file.Write(encodedItems)
	file.Close()
}

func updateFile(fileName string, items []Item) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}

	encodedItems, _ := json.Marshal(items)

	file.Truncate(0)
	file.Seek(0, 0)

	file.Write(encodedItems)
	file.Close()
}

func listItems(fileName string, writer io.Writer) (err error) {

	_, byteValue := readFromFile(fileName)
	writer.Write(byteValue)
	return err
}

func findItemByID(fileName string, id string) (item Item) {

	items, _ := readFromFile(fileName)
	for i := 0; i < len(items); i++ {
		if items[i].Id == id {
			item = items[i]
			break
		}
	}

	return item
}

func addItem(item string, fileName string) (err error) {

	var newItem Item

	id, email, age := parseItem(item)
	existedItem := findItemByID(fileName, id)

	if existedItem != (Item{}) {
		err := fmt.Errorf("Item with id %s already exists", existedItem.Id)
		return err
	}

	newItem.Id, newItem.Email, newItem.Age = id, email, age
	items, _ := readFromFile(fileName)

	newItems := append(items, newItem)
	writeToFile(fileName, newItems)

	return nil
}

func removeItem(fileName string, id string, writer io.Writer) (err error) {

	item := findItemByID(fileName, id)
	if item == (Item{}) {

		_, err = writer.Write([]byte(fmt.Sprintf("Item with id %s not found", id)))

		return err

	} else {

		items, _ := readFromFile(fileName)

		for i := 0; i < len(items); i++ {
			if items[i].Id == id {
				items = append(items[:i], items[i+1:]...)
				break
			}
		}

		updateFile(fileName, items)

		return nil
	}
}

func Perform(args Arguments, writer io.Writer) (err error) {

	// decompose args into veriables
	operation := args["operation"]
	item := strings.Trim(args["item"], "{}")
	fileName := args["fileName"]
	id := args["id"]

	if fileName == "" {
		return errors.New("-fileName flag has to be specified")
	}

	switch operation {
	case "":
		err = errors.New("-operation flag has to be specified")
	case "list":
		err = listItems(fileName, writer)
		if err != nil {
			return err
		}
	case "findById":
		if id == "" {
			return errors.New("-id flag has to be specified")
		}

		item := findItemByID(fileName, id)
		if item.Id == "" {
			writer.Write([]byte(""))
		} else {
			encodedItem, err := json.Marshal(item)

			if err != nil {
				return err
			}

			writer.Write(encodedItem)
		}
	case "add":
		if item == "" {
			return errors.New("-item flag has to be specified")
		}

		err = addItem(item, fileName)

		if err != nil {
			fmt.Fprintf(writer, string(err.Error()))
			return nil
		}
	case "remove":
		if id == "" {
			return errors.New("-id flag has to be specified")
		}

		removeItem(fileName, id, writer)
	default:
		err = fmt.Errorf("Operation %s not allowed!", operation)
	}

	return err
}

func main() {

	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}

func parseArgs() Arguments {
	var operation, id, item, fileName string

	flag.StringVar(&operation, "operation", "", "operation flag")
	flag.StringVar(&item, "item", "", "item flag")
	flag.StringVar(&id, "id", "", "item flag")
	flag.StringVar(&fileName, "flagName", "", "fileName flag")
	flag.Parse()

	return Arguments{"operation": operation, "id": id, "item": item, "fileName": fileName}
}
