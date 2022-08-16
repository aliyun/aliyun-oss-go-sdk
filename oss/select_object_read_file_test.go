package oss

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func handleError(err error) error {
	if err == nil {
		return nil
	}
	return err
}

func readCsvLine(fileName string) (int, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	rd := csv.NewReader(file)
	rc, err := rd.ReadAll()
	return len(rc), err
}
func readCsvIsEmpty(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()
	var out string
	var i, index int
	var indexYear, indexStateAbbr, indexCityName, indexPopulationCount int

	rd := bufio.NewReader(file)
	for {
		line, err := rd.ReadString('\n') // read a line
		if io.EOF == err {
			break
		}
		if err != nil {
			return "", err
		}

		sptLint := strings.Split(line, ",")
		if i == 0 {
			i = 1
			for _, val := range sptLint {
				switch val {
				case "Year":
					indexYear = index
				case "StateAbbr":
					indexStateAbbr = index
				case "CityName":
					indexCityName = index
				case "PopulationCount":
					indexPopulationCount = index
				}
				index++
			}
		} else {
			if sptLint[indexCityName] != "" {
				outLine := sptLint[indexYear] + "," + sptLint[indexStateAbbr] + "," + sptLint[indexCityName] + "," + sptLint[indexPopulationCount] + "\n"
				out += outLine
			}
		}
	}

	return out, nil
}

func readCsvLike(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()
	var out string
	var i, index int
	var indexYear, indexStateAbbr, indexCityName, indexPopulationCount, indexMeasure int

	rd := bufio.NewReader(file)
	for {
		line, err := rd.ReadString('\n') // read a line
		if io.EOF == err {
			break
		}
		if err != nil {
			return "", err
		}

		//utf8Lint := ConvertToString(line,"gbk", "utf-8")
		sptLint := strings.Split(line[:(len(line)-1)], ",")
		if i == 0 {
			i = 1
			for _, val := range sptLint {
				switch val {
				case "Year":
					indexYear = index
				case "StateAbbr":
					indexStateAbbr = index
				case "CityName":
					indexCityName = index
				case "Short_Question_Text":
					indexPopulationCount = index
				case "Measure":
					indexMeasure = index
				}
				index++
			}
		} else {
			if sptLint[indexMeasure] != "" {
				reg := regexp.MustCompile("^.*blood pressure.*Years$")
				res := reg.FindAllString(sptLint[indexMeasure], -1)
				if len(res) > 0 {
					outLine := sptLint[indexYear] + "," + sptLint[indexStateAbbr] + "," + sptLint[indexCityName] + "," + sptLint[indexPopulationCount] + "\n"
					out += outLine
				}
			}
		}
	}

	return out, nil
}

func readCsvRange(fileName string, l int, r int) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()
	var out string
	var i, index int
	var indexYear, indexStateAbbr, indexCityName, indexPopulationCount int

	rd := bufio.NewReader(file)
	for j := 0; j < r+1; j++ {
		if j < l {
			continue
		}
		line, err := rd.ReadString('\n') // read a line
		if io.EOF == err {
			break
		}
		if err != nil {
			return "", err
		}

		sptLint := strings.Split(line[:(len(line)-1)], ",")
		if i == 0 {
			i = 1
			for _, val := range sptLint {
				switch val {
				case "Year":
					indexYear = index
				case "StateAbbr":
					indexStateAbbr = index
				case "CityName":
					indexCityName = index
				case "Short_Question_Text":
					indexPopulationCount = index
				}
				index++
			}
		} else {
			outLine := sptLint[indexYear] + "," + sptLint[indexStateAbbr] + "," + sptLint[indexCityName] + "," + sptLint[indexPopulationCount] + "\n"
			out += outLine
		}
	}

	return out, nil
}

func readCsvFloatAgg(fileName string) (avg, max, sum float64, er error) {
	file, err := os.Open(fileName)
	if err != nil {
		er = err
		return
	}
	defer file.Close()
	var i, index int
	var indexDataValue int

	rd := csv.NewReader(file)

	for {
		rc, err := rd.Read()
		if io.EOF == err {
			break
		}
		if err != nil {
			er = err
			return
		}
		if i == 0 {
			i = 1
			for index = 0; index < len(rc); index++ {
				if rc[index] == "Data_Value" {
					indexDataValue = index
				}
			}
		} else {
			if rc[indexDataValue] != "" {
				s1, err := strconv.ParseFloat(rc[indexDataValue], 64)
				if err != nil {
					er = err
					return
				}
				sum += s1
				if s1 > max {
					max = s1
				}
				i++
			}
		}
	}
	avg = sum / float64(i-1)
	return
}
func readCsvConcat(fileName string) (string, error) {
	var out string
	file, err := os.Open(fileName)
	if err != nil {
		return out, err
	}
	defer file.Close()
	var i int
	var indexDataValue int
	var indexYear, indexStateAbbr, indexCityName, indexShortQuestionText, indexDataValueUnit int

	rd := csv.NewReader(file)

	for {
		rc, err := rd.Read()
		if io.EOF == err {
			break
		}
		if err != nil {
			return out, err
		}
		if i == 0 {
			for j, v := range rc {
				switch v {
				case "Year":
					indexYear = j
				case "StateAbbr":
					indexStateAbbr = j
				case "CityName":
					indexCityName = j
				case "Short_Question_Text":
					indexShortQuestionText = j
				case "Data_Value_Unit":
					indexDataValueUnit = j
				case "Data_Value":
					indexDataValue = j
				}
			}
		} else {
			i++
			if rc[indexDataValue] != "" || rc[indexDataValueUnit] != "" {
				reg := regexp.MustCompile("^14.8.*$")
				reD := reg.FindAllString(rc[indexDataValue], -1)
				reDU := reg.FindAllString(rc[indexDataValueUnit], -1)
				if len(reD) > 0 || len(reDU) > 0 {
					outLine := rc[indexYear] + "," + rc[indexStateAbbr] + "," + rc[indexCityName] + "," + rc[indexShortQuestionText] + "\n"
					out += outLine
				}
			}
		}
		i++
	}
	return out, nil
}
func readCsvComplicateCondition(fileName string) (string, error) {
	var out string
	file, err := os.Open(fileName)
	if err != nil {
		return out, err
	}
	defer file.Close()
	var i int
	var indexDataValue, indexCategory, indexHighConfidenceLimit, indexMeasure int
	var indexYear, indexStateAbbr, indexCityName, indexShortQuestionText, indexDataValueUnit int

	rd := csv.NewReader(file)

	for {
		rc, err := rd.Read()
		if io.EOF == err {
			break
		}
		if err != nil {
			return out, err
		}
		if i == 0 {
			for j, v := range rc {
				switch v {
				case "Year":
					indexYear = j
				case "StateAbbr":
					indexStateAbbr = j
				case "CityName":
					indexCityName = j
				case "Short_Question_Text":
					indexShortQuestionText = j
				case "Data_Value_Unit":
					indexDataValueUnit = j
				case "Data_Value":
					indexDataValue = j
				case "Measure":
					indexMeasure = j
				case "Category":
					indexCategory = j
				case "High_Confidence_Limit":
					indexHighConfidenceLimit = j
				}
			}
		} else {
			reg := regexp.MustCompile("^.*18 Years$")
			reM := reg.FindAllString(rc[indexMeasure], -1)
			var dataV, limitV float64
			if rc[indexDataValue] != "" {
				dataV, err = strconv.ParseFloat(rc[indexDataValue], 64)
				if err != nil {
					return out, err
				}
			}
			if rc[indexHighConfidenceLimit] != "" {
				limitV, err = strconv.ParseFloat(rc[indexHighConfidenceLimit], 64)
				if err != nil {
					return out, err
				}
			}
			if dataV > 14.8 && rc[indexDataValueUnit] == "%" || len(reM) > 0 &&
				rc[indexCategory] == "Unhealthy Behaviors" || limitV > 70.0 {
				outLine := rc[indexYear] + "," + rc[indexStateAbbr] + "," + rc[indexCityName] + "," + rc[indexShortQuestionText] + "," + rc[indexDataValue] + "," + rc[indexDataValueUnit] + "," + rc[indexCategory] + "," + rc[indexHighConfidenceLimit] + "\n"
				out += outLine
			}
		}
		i++
	}
	return out, nil
}

type Extra struct {
	Address     string `json:"address"`
	ContactForm string `json:"contact_form"`
	Fax         string `json:"fax,omitempty"`
	How         string `json:"how,omitempty"`
	Office      string `json:"office"`
	RssUrl      string `json:"rss_url,omitempty"`
}

type Person struct {
	Bioguideid  string  `json:"bioguideid"`
	Birthday    string  `json:"birthday"`
	Cspanid     int     `json:"cspanid"`
	Firstname   string  `json:"firstname"`
	Gender      string  `json:"gender"`
	GenderLabel string  `json:"gender_label"`
	Lastname    string  `json:"lastname"`
	Link        string  `json:"link"`
	Middlename  string  `json:"middlename"`
	Name        string  `json:"name"`
	Namemod     string  `json:"namemod"`
	Nickname    string  `json:"nickname"`
	Osid        string  `json:"osid"`
	Pvsid       *string `json:"pvsid"`
	Sortname    string  `json:"sortname"`
	Twitterid   *string `json:"twitterid"`
	Youtubeid   *string `json:"youtubeid"`
}

type JsonLineSt struct {
	Caucus            *string `json:"caucus"`
	CongressNumbers   []int   `json:"congress_numbers"`
	Current           bool    `json:"current"`
	Description       string  `json:"description"`
	District          *string `json:"district"`
	Enddate           string  `json:"enddate"`
	Extra             Extra   `json:"extra"`
	LeadershipTitle   *string `json:"leadership_title"`
	Party             string  `json:"party"`
	Person            Person  `json:"person"`
	Phone             string  `json:"phone"`
	RoleType          string  `json:"role_type"`
	RoleTypeLabel     string  `json:"role_type_label"`
	SenatorClass      string  `json:"senator_class"`
	SenatorClassLabel string  `json:"senator_class_label"`
	SenatorRank       string  `json:"senator_rank"`
	SenatorRankLabel  string  `json:"senator_rank_label"`
	Startdate         string  `json:"startdate"`
	State             string  `json:"state"`
	Title             string  `json:"title"`
	TitleLong         string  `json:"title_long"`
	Website           string  `json:"website"`
}
type Metast struct {
	limit      int
	Offset     int
	TotalCount int
}

type JsonSt struct {
	Meta    Metast
	Objects []JsonLineSt `json:"objects"`
}

func readJsonDocument(fileName string) (string, error) {
	var out string
	var data JsonSt
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	for _, v := range data.Objects {
		if v.Party == "Democrat" {
			lint, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			lints := strings.Replace(string(lint), "\\u0026", "&", -1)
			out += lints + ","
		}
	}

	return out, err
}
func readJsonLinesLike(fileName string) (string, error) {
	var out string
	var data JsonSt
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	reg := regexp.MustCompile("^1959.*")
	for _, v := range data.Objects {
		reB := reg.FindAllString(v.Person.Birthday, -1)
		if len(reB) > 0 {
			lints := "{\"firstname\":\"" + v.Person.Firstname + "\",\"lastname\":\"" + v.Person.Lastname + "\"}"
			out += lints + ","
		}
	}

	return out, err
}

func readJsonLinesRange(fileName string, l, r int) (string, error) {
	var out string
	var data JsonSt
	var i int
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	for _, v := range data.Objects {
		if i < l {
			continue
		}
		if i >= r {
			break
		}
		extrb, err := json.Marshal(v.Extra)
		if err != nil {
			return "", err
		}
		extr := strings.Replace(string(extrb), "\\u0026", "&", -1)

		lints := "{\"firstname\":\"" + v.Person.Firstname + "\",\"lastname\":\"" + v.Person.Lastname +
			"\",\"extra\":" + extr + "}"
		out += lints + ","
		i++
	}

	return out, err
}

func readJsonFloatAggregation(fileName string) (float64, float64, float64, error) {
	var avg, max, min, sum float64
	var data JsonSt
	var i int
	file, err := os.Open(fileName)
	if err != nil {
		return avg, max, min, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	for _, v := range data.Objects {
		if i == 0 {
			min = float64(v.Person.Cspanid)
		}
		if max < float64(v.Person.Cspanid) {
			max = float64(v.Person.Cspanid)
		}
		if min > float64(v.Person.Cspanid) {
			min = float64(v.Person.Cspanid)
		}
		sum += float64(v.Person.Cspanid)
		i++
	}
	avg = sum / float64(i)
	return avg, max, min, err
}

func readJsonDocumentConcat(fileName string) (string, error) {
	var out string
	var data JsonSt
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)

	for _, v := range data.Objects {
		if v.Person.Firstname+v.Person.Lastname == "JohnKennedy" {
			extrb, err := json.Marshal(v.Person)
			if err != nil {
				return "", err
			}
			extr := "{\"person\":" + strings.Replace(string(extrb), "\\u0026", "&", -1) + "}"
			out += extr + ","
		}
	}

	return out, err
}

func readJsonComplicateConcat(fileName string) (string, error) {
	var out string
	var data JsonSt
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)

	for _, v := range data.Objects {
		if v.Startdate > "2017-01-01" && v.SenatorRank == "junior" ||
			v.State == "CA" && v.Party == "Repulican" {
			cn := "["
			for _, vv := range v.CongressNumbers {
				cn += strconv.Itoa(vv) + ","
			}
			cn = cn[:len(cn)-1] + "]"
			lints := "{\"firstname\":\"" + v.Person.Firstname + "\",\"lastname\":\"" + v.Person.Lastname + "\",\"congress_numbers\":" + cn + "}"
			out += lints + ","
		}
	}

	return out, err
}
