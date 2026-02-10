package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds all AurumFlow configuration.
type Config struct {
	API           APIConfig           `json:"api" yaml:"api"`
	Risk          RiskConfig          `json:"risk" yaml:"risk"`
	Indicators    IndicatorsConfig    `json:"indicators" yaml:"indicators"`
	Strategy      StrategyConfig      `json:"strategy" yaml:"strategy"`
	Timeframes    TimeframesConfig    `json:"timeframes" yaml:"timeframes"`
	Logging       LoggingConfig       `json:"logging" yaml:"logging"`
	DataQuality   *DataQualityConfig  `json:"data_quality" yaml:"data_quality"`
	M5Refiner     *M5RefinerConfig    `json:"m5_refiner" yaml:"m5_refiner"`
	Notifications *NotificationsConfig `json:"notifications" yaml:"notifications"`
}

// NotificationsConfig holds notification subsystem settings (Pushover, Telegram). Optional; when nil or enabled=false, no notifications are sent.
type NotificationsConfig struct {
	Enabled    bool                  `json:"enabled" yaml:"enabled"`
	Provider   string                `json:"provider" yaml:"provider"` // "pushover" (default)
	Pushover   *PushoverConfig       `json:"pushover" yaml:"pushover"`
	Telegram   *TelegramConfig       `json:"telegram" yaml:"telegram"`
	Categories *NotifCategoriesConfig `json:"categories" yaml:"categories"`
	RateLimit  *NotifRateLimitConfig `json:"rate_limit" yaml:"rate_limit"`
	Thresholds *NotifThresholdsConfig `json:"thresholds" yaml:"thresholds"`
	Summaries  *NotifSummariesConfig `json:"summaries" yaml:"summaries"`
	QueueCap   int                   `json:"queue_cap" yaml:"queue_cap"` // bounded channel capacity (default 100)
	InstanceID string                `json:"instance_id" yaml:"instance_id"` // optional; empty = use epic in main
	GlobalDedup *GlobalDedupConfig   `json:"global_dedup" yaml:"global_dedup"` // optional; when set, POSITION_OPENED/CLOSED dedup by deal_ref (e.g. Redis)
	HeartbeatIntervalMinutes int    `json:"heartbeat_interval_minutes" yaml:"heartbeat_interval_minutes"` // optional; 0 = use default (5 ticks / 10 min)
	HeartbeatAlignTop        bool   `json:"heartbeat_align_top" yaml:"heartbeat_align_top"`               // optional; when true, align to :00 and :30 (if interval 30) or top of hour
}

// GlobalDedupConfig holds settings for global dedup backend (Redis, etc.).
type GlobalDedupConfig struct {
	Type     string             `json:"type" yaml:"type"`         // "redis" or empty
	Redis    *RedisDedupConfig  `json:"redis" yaml:"redis"`       // required when type=redis
	TTLDays  int                `json:"ttl_days" yaml:"ttl_days"` // key TTL in days (default 7)
}

// RedisDedupConfig holds Redis connection for global dedup.
type RedisDedupConfig struct {
	Addr     string `json:"addr" yaml:"addr"`           // e.g. "localhost:6379"
	Password string `json:"password" yaml:"password"`   // optional
	DB       int    `json:"db" yaml:"db"`              // DB index (default 0)
}

// TelegramConfig holds Telegram Bot API settings for channel/group notifications.
// For public channels the bot must be added as admin. ChatID can be @channelname or numeric id.
// Token/chat_id can be overridden via AURUMFLOW_TELEGRAM_BOT_TOKEN / AURUMFLOW_TELEGRAM_CHAT_ID.
type TelegramConfig struct {
	Enabled             bool     `json:"enabled" yaml:"enabled"`
	BotToken            string   `json:"bot_token" yaml:"bot_token"`
	ChatID              string   `json:"chat_id" yaml:"chat_id"`
	ParseMode           string   `json:"parse_mode" yaml:"parse_mode"` // "HTML" (default) or "MarkdownV2"
	Events              []string `json:"events" yaml:"events"`         // optional; if empty, send all events that pass Policy
	IncludeMarketContext bool    `json:"include_market_context" yaml:"include_market_context"`
	FailPolicy          string   `json:"fail_policy" yaml:"fail_policy"` // "best_effort" (recommended)
	RateLimitPerMin     int      `json:"rate_limit_per_min" yaml:"rate_limit_per_min"` // 0 = no extra limit
}

// PushoverConfig holds Pushover API settings. Token and User are required for sending; can be overridden via AURUMFLOW_PUSHOVER_TOKEN / AURUMFLOW_PUSHOVER_USER.
type PushoverConfig struct {
	APIURL                 string `json:"api_url" yaml:"api_url"`
	Token                  string `json:"token" yaml:"token"`
	User                   string `json:"user" yaml:"user"`
	Device                 string `json:"device" yaml:"device"`
	Sound                  string `json:"sound" yaml:"sound"`
	DefaultPriority        int    `json:"default_priority" yaml:"default_priority"`
	EmergencyRetrySeconds  int    `json:"emergency_retry_seconds" yaml:"emergency_retry_seconds"`
	EmergencyExpireSeconds int    `json:"emergency_expire_seconds" yaml:"emergency_expire_seconds"`
}

// NotifCategoriesConfig toggles notification categories (health, market, signals, execution, risk, summary).
type NotifCategoriesConfig struct {
	Health    bool `json:"health" yaml:"health"`
	Market    bool `json:"market" yaml:"market"`
	Signals   bool `json:"signals" yaml:"signals"`
	Execution bool `json:"execution" yaml:"execution"`
	Risk      bool `json:"risk" yaml:"risk"`
	Summary   bool `json:"summary" yaml:"summary"`
}

// NotifRateLimitConfig holds per-event throttle and dedup/aggregate windows.
type NotifRateLimitConfig struct {
	MinIntervalSecondsByEvent map[string]int `json:"min_interval_seconds_by_event" yaml:"min_interval_seconds_by_event"`
	DedupWindowSeconds        int            `json:"dedup_window_seconds" yaml:"dedup_window_seconds"`
	AggregateWindowSeconds   int            `json:"aggregate_window_seconds" yaml:"aggregate_window_seconds"`
}

// NotifThresholdsConfig holds thresholds for DD, PnL, and SWEEP_CONFIRMED filter (SweepMinStrength).
type NotifThresholdsConfig struct {
	NotifyOnDDPct       float64  `json:"notify_on_dd_pct" yaml:"notify_on_dd_pct"`
	NotifyOnPnlR        []float64 `json:"notify_on_pnl_r" yaml:"notify_on_pnl_r"`
	NotifyOnPnlPct      []float64 `json:"notify_on_pnl_pct" yaml:"notify_on_pnl_pct"`
	NotifyWhenNearSLATR float64  `json:"notify_when_near_sl_atr" yaml:"notify_when_near_sl_atr"`
	SweepMinStrength     float64  `json:"sweep_min_strength" yaml:"sweep_min_strength"` // SWEEP_CONFIRMED: notify only if strength >= this, or zoneBuy/zoneSell != 0, or type changed
}

// NotifSummariesConfig toggles hourly/session/daily summaries (Phase 2).
type NotifSummariesConfig struct {
	Hourly  bool `json:"hourly" yaml:"hourly"`
	Session bool `json:"session" yaml:"session"`
	Daily   bool `json:"daily" yaml:"daily"`
}

// M5RefinerConfig holds M5 timing refiner settings (soft timing in READY only).
type M5RefinerConfig struct {
	Enabled                   bool   `json:"enabled" yaml:"enabled"`                                           // use M5 for timing in READY (default true)
	DegradeIfUnavailable      bool   `json:"degrade_if_unavailable" yaml:"degrade_if_unavailable"`             // if true, continue M15-only when M5 stale (default true)
	M5MaxAgeMinutes           int    `json:"m5_max_age_minutes" yaml:"m5_max_age_minutes"`                     // max age of last M5 candle in minutes (default 10)
	M5CoherenceDiffMinutes    int    `json:"m5_coherence_diff_minutes" yaml:"m5_coherence_diff_minutes"`       // max diff M15 last - M5 last in minutes (default 15)
	M5ScoreWeight             int    `json:"m5_score_weight" yaml:"m5_score_weight"`                           // weight for m5Score +1/-1/0 (default 1)
	M5TimingMode               string `json:"m5_timing_mode" yaml:"m5_timing_mode"`                           // SOFT only for now
	M5TimingRejectAction       string `json:"m5_timing_reject_action" yaml:"m5_timing_reject_action"`           // RECYCLE_TO_WAIT_PULLBACK
	M5RecycleTTLCandles       int    `json:"m5_recycle_ttl_candles" yaml:"m5_recycle_ttl_candles"`             // M15 candles TTL after recycle (default 5)
}

// DataQualityConfig holds data freshness and reopen warmup settings (production readiness).
type DataQualityConfig struct {
	MaxAgeM15Minutes        int  `json:"max_age_m15_minutes" yaml:"max_age_m15_minutes"`               // max age of last M15 candle in minutes (default 20)
	MaxAgeH1Minutes         int  `json:"max_age_h1_minutes" yaml:"max_age_h1_minutes"`                 // max age of last H1 candle in minutes (default 75)
	MaxAgeH4Minutes         int  `json:"max_age_h4_minutes" yaml:"max_age_h4_minutes"`                  // max age of last H4 candle in minutes (default 300)
	ReopenWarmupM15Candles  int  `json:"reopen_warmup_m15_candles" yaml:"reopen_warmup_m15_candles"`  // candles to wait after market reopen before evaluating (default 3)
	BlockOnDataFrozen       bool `json:"block_on_data_frozen" yaml:"block_on_data_frozen"`            // if true, skip signal evaluation when data is stale (default true)
}

// LoggingConfig holds log file and journal settings.
type LoggingConfig struct {
	LogDir       string `json:"log_dir" yaml:"log_dir"`             // directory for per-run log files; empty = disabled; default "logs"
	JournalDir   string `json:"journal_dir" yaml:"journal_dir"`     // directory for journal JSONL; default "journals"
	JournalEnabled bool `json:"journal_enabled" yaml:"journal_enabled"` // if true, write setup/signal/order/position events to journal
}

// Environment mode: "demo" or "live".
const (
	ModeDemo = "demo"
	ModeLive = "live"
)

// APIConfig holds Capital.com API settings.
type APIConfig struct {
	Mode       string  `json:"mode" yaml:"mode"`             // "demo" or "live"; if set, BaseURL is derived
	BaseURL    string  `json:"api_base_url" yaml:"api_base_url"`
	APIKey     string  `json:"api_key" yaml:"api_key"`
	Identifier string  `json:"identifier" yaml:"identifier"`
	Password   string  `json:"password" yaml:"password"`
	AccountID  string  `json:"account_id" yaml:"account_id"` // optional; account to operate (switch after login)
	MaxSpread  float64 `json:"max_spread" yaml:"max_spread"` // optional; skip order if spread > this (0 = disabled)
}

// RiskConfig holds risk management parameters.
// ValuePerPoint: money per 1.0 price movement per 1.0 size (stopDistance is in price units). Default 1.0 for XAUUSD 1:1; set from broker/API if available.
type RiskConfig struct {
	RiskPerTrade       float64 `json:"risk_per_trade" yaml:"risk_per_trade"`
	MaxTrades          int     `json:"max_trades" yaml:"max_trades"`
	DailyDrawdownLimit float64 `json:"daily_drawdown_limit" yaml:"daily_drawdown_limit"`
	ValuePerPoint      float64 `json:"value_per_point" yaml:"value_per_point"` // money per 1.0 price move per 1.0 size; 0 or negative => 1.0
}

// IndicatorsConfig holds RSI and ATR parameters.
type IndicatorsConfig struct {
	RSIPeriod    int     `json:"rsi_period" yaml:"rsi_period"`
	ATRPeriod    int     `json:"atr_period" yaml:"atr_period"`
	MinATR       float64 `json:"min_atr" yaml:"min_atr"`
	MaxATR       float64 `json:"max_atr" yaml:"max_atr"`
	UseRSIWilder bool    `json:"use_rsi_wilder" yaml:"use_rsi_wilder"` // if true, use Wilder smoothing (matches TradingView)
}

// StrategyConfig holds strategy parameters.
type StrategyConfig struct {
	EntryDelayCandles int      `json:"entry_delay_candles" yaml:"entry_delay_candles"`
	SwingLookback     int      `json:"swing_lookback" yaml:"swing_lookback"`
	ScoreThreshold    int      `json:"score_threshold" yaml:"score_threshold"`
	RSIRequired       bool     `json:"rsi_required" yaml:"rsi_required"` // deprecated: use rsi_mode; if rsi_mode empty, true = GATE, false = BOOST
	RSIMode           string   `json:"rsi_mode" yaml:"rsi_mode"`         // GATE (reject if not in zone), BOOST (+1 in zone), PENALTY (+2/0/-1); default BOOST
	RSIBuyLow         float64  `json:"rsi_buy_low" yaml:"rsi_buy_low"`   // buy zone low (default 38)
	RSIBuyHigh        float64  `json:"rsi_buy_high" yaml:"rsi_buy_high"` // buy zone high (default 42)
	RSISellLow        float64  `json:"rsi_sell_low" yaml:"rsi_sell_low"` // sell zone low (default 58)
	RSISellHigh       float64  `json:"rsi_sell_high" yaml:"rsi_sell_high"` // sell zone high (default 62)
	TradingSessions   []string `json:"trading_sessions" yaml:"trading_sessions"` // allowed sessions: ["LONDON", "NY", "ASIA"] or ["ALL"] (default: ["ALL"])
	BlockLondonNYOverlap *bool `json:"block_london_ny_overlap" yaml:"block_london_ny_overlap"` // if true or unset, skip signal evaluation during LONDON+NY overlap; set false to allow
	// Sweep detection (institutional: broke zone + wick + close inside)
	SweepCloseTolerance float64 `json:"sweep_close_tolerance" yaml:"sweep_close_tolerance"` // ATR multiplier for close-inside, default 0.2
	SweepMinPenetration float64 `json:"sweep_min_penetration" yaml:"sweep_min_penetration"` // min wick in ATR units, default 0.12
	SweepMinWickRatio   float64 `json:"sweep_min_wick_ratio" yaml:"sweep_min_wick_ratio"`   // min wick/body ratio, default 1.4
	SweepDebugLog       bool    `json:"sweep_debug_log" yaml:"sweep_debug_log"`             // log [LiquidityEngine] and [SweepAnalyzer] when true
	// H1 bias filter (institutional: only trade in direction of H1 trend)
	UseH1Filter       bool `json:"use_h1_filter" yaml:"use_h1_filter"`             // if true, BUY only when H1 BULLISH, SELL only when H1 BEARISH
	BlockH1Range      bool `json:"block_h1_range" yaml:"block_h1_range"`           // if true and UseH1Filter, do not trade when H1 trend is RANGE (default true)
	Crypto            bool `json:"crypto" yaml:"crypto"`                         // if true, allow H1 RANGE when finalScore >= threshold + h1_range_extra_score (default false)
	H1RangeExtraScore int  `json:"h1_range_extra_score" yaml:"h1_range_extra_score"` // extra score required for H1 RANGE when crypto=true (default 1)
	H1SwingLookback   int  `json:"h1_swing_lookback" yaml:"h1_swing_lookback"`   // swing lookback for H1 structure (default 5)
	// H4 regime filter (institutional: do not trade against H4; gold is macro-driven)
	UseH4Filter       bool `json:"use_h4_filter" yaml:"use_h4_filter"`             // if true, BUY only when H4 BULLISH, SELL only when H4 BEARISH
	BlockH4Range      bool `json:"block_h4_range" yaml:"block_h4_range"`           // if true and UseH4Filter, do not trade when H4 trend is RANGE
	H4SwingLookback   int  `json:"h4_swing_lookback" yaml:"h4_swing_lookback"`   // swing lookback for H4 structure (default 5)
	// State machine TTL and invalidation (reduce late entries)
	StateTTLCandles             int  `json:"state_ttl_candles" yaml:"state_ttl_candles"`                         // max M15 candles in SEEK_LIQUIDITY/WAIT_SWEEP/WAIT_PULLBACK/READY before reset to IDLE (default 12)
	InvalidateOnOppositeStructure bool `json:"invalidate_on_opposite_structure" yaml:"invalidate_on_opposite_structure"` // if true, reset to IDLE when structure trend opposes setup direction (default true)
	// Exit management (BE at 1R, partials)
	UseBreakEven bool    `json:"use_break_even" yaml:"use_break_even"`   // if true, move SL to entry when price reaches BreakEvenR
	BreakEvenR   float64 `json:"break_even_r" yaml:"break_even_r"`       // R level to trigger BE (e.g. 1.0)
	UsePartial   bool    `json:"use_partial" yaml:"use_partial"`         // if true, close part at PartialR and let rest run to RunnerR
	PartialR     float64 `json:"partial_r" yaml:"partial_r"`            // R level for partial close (e.g. 1.0)
	RunnerR      float64 `json:"runner_r" yaml:"runner_r"`               // R target for remainder (e.g. 2.0 or 3.0)
}

// TimeframesConfig holds timeframe names (mapped to API resolutions).
type TimeframesConfig struct {
	H4  string `json:"h4" yaml:"h4"`   // HOUR_4 (regime)
	H1  string `json:"h1" yaml:"h1"`   // HOUR (trend)
	M15 string `json:"m15" yaml:"m15"` // MINUTE_15 (setup)
	M5  string `json:"m5" yaml:"m5"`   // MINUTE_5 (trigger, optional)
}

// dataQualityDefaults returns default DataQualityConfig when section is missing.
func dataQualityDefaults() DataQualityConfig {
	return DataQualityConfig{
		MaxAgeM15Minutes:       20,
		MaxAgeH1Minutes:        75,
		MaxAgeH4Minutes:        300,
		ReopenWarmupM15Candles: 3,
		BlockOnDataFrozen:      true,
	}
}

// Load reads config from a JSON or YAML file and validates required fields.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var c Config
	switch {
	case strings.HasSuffix(strings.ToLower(path), ".yaml") || strings.HasSuffix(strings.ToLower(path), ".yml"):
		if err := yaml.Unmarshal(data, &c); err != nil {
			return nil, fmt.Errorf("parse config yaml: %w", err)
		}
	default:
		if err := json.Unmarshal(data, &c); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	c.ApplyEnvOverrides()
	return &c, nil
}

// ApplyEnvOverrides fills API secrets from environment if not set in config.
func (c *Config) ApplyEnvOverrides() {
	if v := os.Getenv("AURUMFLOW_API_KEY"); v != "" && c.API.APIKey == "" {
		c.API.APIKey = v
	}
	if v := os.Getenv("AURUMFLOW_IDENTIFIER"); v != "" && c.API.Identifier == "" {
		c.API.Identifier = v
	}
	if v := os.Getenv("AURUMFLOW_PASSWORD"); v != "" && c.API.Password == "" {
		c.API.Password = v
	}
	if v := os.Getenv("AURUMFLOW_ACCOUNT_ID"); v != "" && c.API.AccountID == "" {
		c.API.AccountID = v
	}
	if c.Notifications != nil && c.Notifications.Pushover != nil {
		if v := os.Getenv("AURUMFLOW_PUSHOVER_TOKEN"); v != "" && c.Notifications.Pushover.Token == "" {
			c.Notifications.Pushover.Token = v
		}
		if v := os.Getenv("AURUMFLOW_PUSHOVER_USER"); v != "" && c.Notifications.Pushover.User == "" {
			c.Notifications.Pushover.User = v
		}
	}
	if c.Notifications != nil && c.Notifications.Telegram != nil {
		if v := os.Getenv("AURUMFLOW_TELEGRAM_BOT_TOKEN"); v != "" && c.Notifications.Telegram.BotToken == "" {
			c.Notifications.Telegram.BotToken = v
		}
		if v := os.Getenv("AURUMFLOW_TELEGRAM_CHAT_ID"); v != "" && c.Notifications.Telegram.ChatID == "" {
			c.Notifications.Telegram.ChatID = v
		}
	}
}

// Validate checks required fields and sets defaults.
func (c *Config) Validate() error {
	// Derive BaseURL from mode if set
	switch strings.ToLower(strings.TrimSpace(c.API.Mode)) {
	case ModeDemo:
		c.API.Mode = ModeDemo
		if c.API.BaseURL == "" {
			c.API.BaseURL = "https://demo-api-capital.backend-capital.com"
		}
	case ModeLive:
		c.API.Mode = ModeLive
		if c.API.BaseURL == "" {
			c.API.BaseURL = "https://api-capital.backend-capital.com"
		}
	default:
		if c.API.Mode != "" {
			return fmt.Errorf("config: api.mode must be %q or %q", ModeDemo, ModeLive)
		}
	}
	// Coherence: if mode is live, URL must not be demo (and vice versa)
	if c.API.Mode == ModeLive && strings.Contains(c.API.BaseURL, "demo-api") {
		return fmt.Errorf("config: api.mode is live but api_base_url points to demo; use https://api-capital.backend-capital.com")
	}
	if c.API.Mode == ModeDemo && c.API.BaseURL != "" && !strings.Contains(c.API.BaseURL, "demo") && strings.Contains(c.API.BaseURL, "capital.com") {
		c.API.BaseURL = "https://demo-api-capital.backend-capital.com"
	}
	if c.API.BaseURL == "" {
		return fmt.Errorf("config: api.api_base_url is required (or set api.mode to demo/live)")
	}
	// Secrets can be provided by env (applied after Load via ApplyEnvOverrides)
	if c.API.APIKey == "" && os.Getenv("AURUMFLOW_API_KEY") == "" {
		return fmt.Errorf("config: api.api_key is required (or set AURUMFLOW_API_KEY)")
	}
	if c.API.Identifier == "" && os.Getenv("AURUMFLOW_IDENTIFIER") == "" {
		return fmt.Errorf("config: api.identifier is required (or set AURUMFLOW_IDENTIFIER)")
	}
	if c.API.Password == "" && os.Getenv("AURUMFLOW_PASSWORD") == "" {
		return fmt.Errorf("config: api.password is required (or set AURUMFLOW_PASSWORD)")
	}
	if c.Risk.RiskPerTrade <= 0 {
		c.Risk.RiskPerTrade = 0.5
	}
	if c.Risk.ValuePerPoint <= 0 {
		c.Risk.ValuePerPoint = 1.0
	}
	if c.Risk.MaxTrades <= 0 {
		c.Risk.MaxTrades = 3
	}
	if c.Risk.DailyDrawdownLimit <= 0 {
		c.Risk.DailyDrawdownLimit = 2.5
	}
	// Data quality: apply defaults if missing or zero
	if c.DataQuality == nil {
		dq := dataQualityDefaults()
		c.DataQuality = &dq
	} else {
		if c.DataQuality.MaxAgeM15Minutes <= 0 {
			c.DataQuality.MaxAgeM15Minutes = 20
		}
		if c.DataQuality.MaxAgeH1Minutes <= 0 {
			c.DataQuality.MaxAgeH1Minutes = 75
		}
		if c.DataQuality.MaxAgeH4Minutes <= 0 {
			c.DataQuality.MaxAgeH4Minutes = 300
		}
		if c.DataQuality.ReopenWarmupM15Candles <= 0 {
			c.DataQuality.ReopenWarmupM15Candles = 3
		}
	}
	if c.M5Refiner == nil {
		c.M5Refiner = &M5RefinerConfig{
			Enabled:                true,
			DegradeIfUnavailable:   true,
			M5MaxAgeMinutes:        10,
			M5CoherenceDiffMinutes: 15,
			M5ScoreWeight:          1,
			M5TimingMode:           "SOFT",
			M5TimingRejectAction:   "RECYCLE_TO_WAIT_PULLBACK",
			M5RecycleTTLCandles:    5,
		}
	} else {
		if c.M5Refiner.M5MaxAgeMinutes <= 0 {
			c.M5Refiner.M5MaxAgeMinutes = 10
		}
		if c.M5Refiner.M5CoherenceDiffMinutes <= 0 {
			c.M5Refiner.M5CoherenceDiffMinutes = 15
		}
		if c.M5Refiner.M5RecycleTTLCandles <= 0 {
			c.M5Refiner.M5RecycleTTLCandles = 5
		}
	}
	if c.Indicators.RSIPeriod <= 0 {
		c.Indicators.RSIPeriod = 9
	}
	if c.Indicators.ATRPeriod <= 0 {
		c.Indicators.ATRPeriod = 14
	}
	if c.Indicators.MinATR <= 0 {
		c.Indicators.MinATR = 80
	}
	if c.Indicators.MaxATR <= 0 {
		c.Indicators.MaxATR = 350
	}
	if c.Strategy.EntryDelayCandles <= 0 {
		c.Strategy.EntryDelayCandles = 3
	}
	if c.Strategy.SwingLookback <= 0 {
		c.Strategy.SwingLookback = 5
	}
	if c.Strategy.ScoreThreshold <= 0 {
		c.Strategy.ScoreThreshold = 6
	}
	if c.Strategy.RSIBuyLow <= 0 {
		c.Strategy.RSIBuyLow = 38
	}
	if c.Strategy.RSIBuyHigh <= 0 {
		c.Strategy.RSIBuyHigh = 42
	}
	if c.Strategy.RSISellLow <= 0 {
		c.Strategy.RSISellLow = 58
	}
	if c.Strategy.RSISellHigh <= 0 {
		c.Strategy.RSISellHigh = 62
	}
	switch strings.ToUpper(strings.TrimSpace(c.Strategy.RSIMode)) {
	case "GATE", "BOOST", "PENALTY":
		c.Strategy.RSIMode = strings.ToUpper(strings.TrimSpace(c.Strategy.RSIMode))
	default:
		if c.Strategy.RSIRequired {
			c.Strategy.RSIMode = "GATE"
		} else {
			c.Strategy.RSIMode = "BOOST"
		}
	}
	if c.Strategy.SweepCloseTolerance <= 0 {
		c.Strategy.SweepCloseTolerance = 0.2
	}
	if c.Strategy.SweepMinPenetration <= 0 {
		c.Strategy.SweepMinPenetration = 0.12
	}
	if c.Strategy.SweepMinWickRatio <= 0 {
		c.Strategy.SweepMinWickRatio = 1.4
	}
	if len(c.Strategy.TradingSessions) == 0 {
		c.Strategy.TradingSessions = []string{"ALL"}
	}
	if c.Strategy.H1SwingLookback <= 0 {
		c.Strategy.H1SwingLookback = 5
	}
	if c.Strategy.H4SwingLookback <= 0 {
		c.Strategy.H4SwingLookback = 5
	}
	if c.Strategy.Crypto && c.Strategy.H1RangeExtraScore <= 0 {
		c.Strategy.H1RangeExtraScore = 1
	}
	if c.Strategy.H1RangeExtraScore > 5 {
		c.Strategy.H1RangeExtraScore = 5
	}
	if c.Strategy.StateTTLCandles <= 0 {
		c.Strategy.StateTTLCandles = 12
	}
	if c.Strategy.BreakEvenR <= 0 {
		c.Strategy.BreakEvenR = 1.0
	}
	if c.Strategy.PartialR <= 0 {
		c.Strategy.PartialR = 1.0
	}
	if c.Strategy.RunnerR <= 0 {
		c.Strategy.RunnerR = 2.0
	}
	if c.Timeframes.H4 == "" {
		c.Timeframes.H4 = "HOUR_4"
	}
	if c.Timeframes.H1 == "" {
		c.Timeframes.H1 = "HOUR"
	}
	if c.Timeframes.M15 == "" {
		c.Timeframes.M15 = "MINUTE_15"
	}
	if c.Timeframes.M5 == "" {
		c.Timeframes.M5 = "MINUTE_5"
	}
	if c.Logging.LogDir == "" {
		c.Logging.LogDir = "logs"
	}
	if c.Logging.JournalDir == "" {
		c.Logging.JournalDir = "journals"
	}
	// Notifications: apply defaults when block is present; do not require token/user when enabled
	if c.Notifications != nil {
		if c.Notifications.Provider == "" {
			c.Notifications.Provider = "pushover"
		}
		if c.Notifications.QueueCap <= 0 {
			c.Notifications.QueueCap = 100
		}
		if c.Notifications.Pushover == nil {
			c.Notifications.Pushover = &PushoverConfig{}
		}
		if c.Notifications.Pushover.APIURL == "" {
			c.Notifications.Pushover.APIURL = "https://api.pushover.net/1/messages.json"
		}
		if c.Notifications.Pushover.EmergencyRetrySeconds <= 0 {
			c.Notifications.Pushover.EmergencyRetrySeconds = 60
		}
		if c.Notifications.Pushover.EmergencyExpireSeconds <= 0 {
			c.Notifications.Pushover.EmergencyExpireSeconds = 3600
		}
		if c.Notifications.Categories == nil {
			c.Notifications.Categories = &NotifCategoriesConfig{
				Health: true, Market: false, Signals: true, Execution: true, Risk: true, Summary: true,
			}
		}
		if c.Notifications.RateLimit == nil {
			c.Notifications.RateLimit = &NotifRateLimitConfig{
				MinIntervalSecondsByEvent: map[string]int{
					"HEARTBEAT":        3600,
					"REGIME_CHANGE":    900,
					"SWEEP_CONFIRMED":  900,
					"SIGNAL_REJECTED":  900,
					"SIGNAL_GENERATED": 0,
					"ORDER_SENT":       0,
					"ORDER_FILLED":     0,
					"POSITION_CLOSED":  0,
					"ERROR":            0,
				},
				DedupWindowSeconds:      900,
				AggregateWindowSeconds: 1800,
			}
		} else {
			if c.Notifications.RateLimit.DedupWindowSeconds <= 0 {
				c.Notifications.RateLimit.DedupWindowSeconds = 900
			}
			if c.Notifications.RateLimit.AggregateWindowSeconds <= 0 {
				c.Notifications.RateLimit.AggregateWindowSeconds = 1800
			}
		}
		if c.Notifications.Thresholds == nil {
			c.Notifications.Thresholds = &NotifThresholdsConfig{}
		}
		if c.Notifications.Summaries == nil {
			c.Notifications.Summaries = &NotifSummariesConfig{}
		}
		if c.Notifications.Telegram != nil && c.Notifications.Telegram.Enabled {
			if c.Notifications.Telegram.ParseMode == "" {
				c.Notifications.Telegram.ParseMode = "HTML"
			}
		}
		if c.Notifications.GlobalDedup != nil {
			if c.Notifications.GlobalDedup.TTLDays <= 0 {
				c.Notifications.GlobalDedup.TTLDays = 7
			}
		}
	}
	return nil
}
