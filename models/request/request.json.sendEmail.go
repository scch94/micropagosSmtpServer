package request

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/scch94/ins_log"
)

type SendEmailRequest struct {
	Utfi                 string `json:"utfi"`
	ServiceType          string `json:"serviceType"`
	OriginNumber         string `json:"origin.number"`
	DestinationNumber    string `json:"destination.number"`
	ValidityPeriod       string `json:"validity_period"`
	ScheduleDeliveryTime string `json:"schedule_delivery_time"`
	ProtocolID           uint8  `json:"protocol_id"`
	EsmeClass            uint8  `json:"esmeClass"`
	PriorityFlag         uint8  `json:"priority_flag"`
	RegisteredDelivery   uint8  `json:"registered_delivery"`
	ReplaceIfPresentFlag uint8  `json:"replace_if_present_flag"`
	Data                 string `json:"data"`
	DataHeaderIndicator  uint8  `json:"data_header_indicator"`
	DataCodingScheme     uint8  `json:"data_coding_scheme"`
	DataLength           uint16 `json:"data_length"`
	MessageType          uint8  `json:"messagetype"`
	TLVTag               int    `json:"TLV_tag"`
	TLVLength            int    `json:"TLV_length"`
	TLVValue             string `json:"TLV_value"`
	Client               string `json:"client"`
}

// metodo para devolver eldata exe como un string
func (request SendEmailRequest) GetMessage(ctx context.Context) (string, error) {

	ctx = ins_log.SetPackageNameInContext(ctx, "request")
	ins_log.Tracef(ctx, "starting to change de hexa string in a normal text - hexa: %v", request.Data)

	//convertir la caeda hexadecimal a bytes
	bytes, err := hex.DecodeString(request.Data)
	if err != nil {
		ins_log.Errorf(ctx, "error when we try to decode de hexa string err: %v", err)
		return "", err
	}

	//convertir los bytes en una cadena de texto
	message := string(bytes)

	return message, nil

}

// metodo que devuelve el origin nummber en formato normal y no internacional
func (request SendEmailRequest) GetDestination() string {
	var destination string
	if strings.HasPrefix(request.DestinationNumber, "598") {
		destination = "0" + request.DestinationNumber[3:]
	} else {
		destination = request.DestinationNumber
	}
	return destination
}
