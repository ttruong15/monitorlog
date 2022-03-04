package monitorLog

var AvailableServices = map[string]string{
	"fail2ban": "/var/log/fail2ban.log",
}

type Processor interface {
	Processing() error
	GenerateReport() string
}
