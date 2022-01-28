package rabbit

type RabbitConfig struct {
	User     string
	Password string
	Host     []string
	Vhost    string
}

func (rC *RabbitConfig) IsNil() bool {
	if rC == nil {
		return true
	}
	return rC.User == "" || rC.Password == "" || len(rC.Host) <= 0
}
