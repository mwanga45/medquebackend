
 package smsB

// import (
// 	"bytes"
// 	"encoding/base64"
// 	"encoding/json"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"

// 	"github.com/brentles-com/brentles_sms/constants"
// )

// func beemApi(info constants.SmsRequest) *constants.AnswerState {
// 	st, id, key, sec := beemKeys()
// 	if st.State != constants.SuccessState {
// 		return st
// 	}

// 	client := &http.Client{}

// 	// build recepient list
// 	sendData := constants.SmsSendRequestBeem{
// 		SourceAddr:   id,
// 		ScheduleTime: "",
// 		Encoding:     0,
// 		Message:      info.Sms,
// 		Recipients:   []constants.SmsRecipientBeem{},
// 	}
// 	for idx, phn := range info.Phone {
// 		sendData.Recipients = append(sendData.Recipients, constants.SmsRecipientBeem{
// 			RecipientId: idx + 1,
// 			DestAddr:    phn,
// 		})
// 	}

// 	jSendData, err := json.Marshal(sendData)
// 	if err != nil {
// 		log.Println(err.Error())
// 		return &constants.AnswerState{
// 			State: constants.ErrorState,
// 			Data:  "Failed to convert data for sending",
// 			Adv:   "none",
// 		}
// 	}

// 	req, er := http.NewRequest("POST", "https://apisms.beem.africa/v1/send", bytes.NewBuffer(jSendData))

// 	if er != nil {
// 		log.Println(er.Error())
// 		return &constants.AnswerState{
// 			State: constants.ErrorState,
// 			Data:  "Failed to setup sms send request",
// 			Adv:   "none",
// 		}
// 	}

// 	req.Header.Add("Content-Type", "application/json")
// 	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(key+":"+sec)))

// 	resp, erP := client.Do(req)
// 	if erP != nil {
// 		log.Println(erP.Error())
// 		return &constants.AnswerState{
// 			State: constants.ErrorState,
// 			Data:  "Failed to send sms request",
// 			Adv:   "none",
// 		}
// 	}
// 	defer resp.Body.Close()

// 	body, erB := io.ReadAll(resp.Body)
// 	if erB != nil {
// 		log.Println(erB.Error())
// 		return &constants.AnswerState{
// 			State: constants.ErrorState,
// 			Data:  "Failed to read body of beem response",
// 			Adv:   "none",
// 		}
// 	}

// 	var dataB constants.BeemSmsSendResponse

// 	erJ := json.Unmarshal(body, &dataB)
// 	if erJ != nil {
// 		log.Println(erJ.Error())
// 		return &constants.AnswerState{
// 			State: constants.ErrorState,
// 			Data:  "Failed to decode beem response",
// 			Adv:   "none",
// 		}
// 	}

// 	if dataB.Code != 100 {
// 		return &constants.AnswerState{
// 			State: constants.ErrorState,
// 			Data:  "Failed to send text due to " + dataB.Message,
// 			Adv:   "none",
// 		}
// 	}

// 	return &constants.AnswerState{
// 		State: constants.SuccessState,
// 		Data:  "Message was successfully sent to receipts",
// 		Adv:   "none",
// 	}

// }

// func beemKeys() (st *constants.AnswerState, id, key, sec string) {
// 	// start with id
// 	idx, exist := os.LookupEnv("BEEM_ID")
// 	if !exist {
// 		return &constants.AnswerState{
// 			State: constants.ErrorState,
// 			Data:  "Failed to get ID",
// 			Adv:   "none",
// 		}, "", "", ""
// 	}

// 	// kye
// 	keyx, exist := os.LookupEnv("BEEM_KEY")
// 	if !exist {
// 		return &constants.AnswerState{
// 			State: constants.ErrorState,
// 			Data:  "Failed to get vendor details",
// 			Adv:   "none",
// 		}, "", "", ""
// 	}

// 	// secret
// 	secret, exist := os.LookupEnv("BEEM_SECRET")
// 	if !exist {
// 		return &constants.AnswerState{
// 			State: constants.ErrorState,
// 			Data:  "Failed to get vendor values",
// 			Adv:   "none",
// 		}, "", "", ""
// 	}

// 	return &constants.AnswerState{
// 		State: constants.SuccessState,
// 		Data:  "success",
// 		Adv:   "none",
// 	}, idx, keyx, secret
// }