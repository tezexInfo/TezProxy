package proxy



type ProxyConfig struct {
	TezosHost       string
	TezosPort       int
	ServerPort      int
	Methods         []string
	ReadTimeout     int
	WriteTimeout    int
	IdleTimeout     int
	RateLimitPeriod int
	RateLimitCount  int64
	Blocked []string
	DontCache []string
	CacheMaxItems int
}