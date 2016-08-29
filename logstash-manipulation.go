func eventTextToJsonString(text string) string {
  text = strings.Replace(text, "\"params\": -", "\"params\": {}", -1) // Replace empty params (e.g GET request)
  text = strings.Replace(text, "\"", "\\x22", -1)
  lineStr, err := strconv.Unquote(fmt.Sprintf(`"%s"`, text))
  if err != nil {
    return ""
  }

  var v map[string]interface{}
  if err := json.Unmarshal([]byte(lineStr), &v); err == nil {
    requestStr, ok := v["request"].(string)
    requestMethodStr, ok2 := v["request_method"].(string)
    if ok && ok2 {
      if requestMethodStr == "GET" {

        // Extract url params
        fromIndex := strings.Index(requestStr, "?")
        lastIndex := strings.Index(requestStr, "HTTP/1.1")
        if fromIndex != -1 && lastIndex != -1 {
          requestParamsStr := requestStr[fromIndex+1:lastIndex-1]
          params := strings.Split(requestParamsStr, "&")
          v["params"] = make(map[string]string)
          for _,element := range params {
            var paramArray []string
            paramArray = strings.Split(element, "=")
            if len(paramArray) == 2 {
				v["params"].(map[string]string)[paramArray[0]] = paramArray[1]
			}
          }
        }

        if fromIndex != -1 {

          // Manipulate the message key
          v["message"] = fmt.Sprintf("%s [%s]", requestStr[0:fromIndex], v["status"])

        } else {

          // Manipulate the message key
          v["message"] = fmt.Sprintf("%s [%s]", strings.Replace(requestStr, " HTTP/1.1", "", -1), v["status"])

        }

      } else {

        // Manipulate the message key
        v["message"] = fmt.Sprintf("%s [%s]", strings.Replace(requestStr, " HTTP/1.1", "", -1), v["status"])

      }

      lineStrBytes, err2 := json.Marshal(&v)
      if err2 == nil {
        lineStr = string(lineStrBytes)
      }
    }

  } else {
    lineStr = ""
  }

  return lineStr
}
