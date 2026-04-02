package common

import "time"

// RetryProfile describes retry behavior for one normalized error class.
type RetryProfile struct {
	Retryable  bool
	MaxRetries int
	Multiplier int
	MinBackoff time.Duration
	MaxBackoff time.Duration
}

// RetryTable maps error classes to profiles.
type RetryTable map[ErrorClass]RetryProfile

// DefaultRetryTable returns profile defaults used by providers.
func DefaultRetryTable() RetryTable {
	return RetryTable{
		ErrorClassAuth:          {Retryable: false, MaxRetries: 0, Multiplier: 1},
		ErrorClassQuota:         {Retryable: false, MaxRetries: 0, Multiplier: 1},
		ErrorClassPermission:    {Retryable: false, MaxRetries: 0, Multiplier: 1},
		ErrorClassModelNotFound: {Retryable: false, MaxRetries: 0, Multiplier: 1},
		ErrorClassInvalidInput:  {Retryable: false, MaxRetries: 0, Multiplier: 1},
		ErrorClassNotFound:      {Retryable: false, MaxRetries: 0, Multiplier: 1},

		ErrorClassRateLimit:   {Retryable: true, MaxRetries: 3, Multiplier: 2, MinBackoff: 250 * time.Millisecond, MaxBackoff: 8 * time.Second},
		ErrorClassTimeout:     {Retryable: true, MaxRetries: 3, Multiplier: 2, MinBackoff: 200 * time.Millisecond, MaxBackoff: 5 * time.Second},
		ErrorClassTransport:   {Retryable: true, MaxRetries: 3, Multiplier: 2, MinBackoff: 200 * time.Millisecond, MaxBackoff: 5 * time.Second},
		ErrorClassUnavailable: {Retryable: true, MaxRetries: 3, Multiplier: 2, MinBackoff: 300 * time.Millisecond, MaxBackoff: 8 * time.Second},
		ErrorClassTransient:   {Retryable: true, MaxRetries: 3, Multiplier: 2, MinBackoff: 250 * time.Millisecond, MaxBackoff: 6 * time.Second},
		ErrorClassCanceled:    {Retryable: false, MaxRetries: 0, Multiplier: 1},
		ErrorClassUnknown:     {Retryable: false, MaxRetries: 0, Multiplier: 1},
	}
}

// Lookup returns the profile for class.
func (t RetryTable) Lookup(class ErrorClass) RetryProfile {
	if p, ok := t[class]; ok {
		if p.Multiplier <= 0 {
			p.Multiplier = 2
		}
		return p
	}
	return RetryProfile{Retryable: false, MaxRetries: 0, Multiplier: 1}
}
