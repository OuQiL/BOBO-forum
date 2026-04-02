package snowflake

import (
	"errors"
	"sync"
	"time"
)

const (
	epoch          = int64(1704067200000)
	nodeIDBits     = uint(10)
	sequenceBits   = uint(12)
	maxNodeID      = int64(-1 ^ (-1 << nodeIDBits))
	maxSequence    = int64(-1 ^ (-1 << sequenceBits))
	nodeIDShift    = sequenceBits
	timestampShift = sequenceBits + nodeIDBits
)

var (
	ErrInvalidNodeID = errors.New("node ID must be between 0 and 1023")
)

type Snowflake struct {
	mu        sync.Mutex
	nodeID    int64
	sequence  int64
	lastTime  int64
}

var defaultNode *Snowflake
var once sync.Once

func Init(nodeID int64) error {
	if nodeID < 0 || nodeID > maxNodeID {
		return ErrInvalidNodeID
	}
	var err error
	once.Do(func() {
		defaultNode = &Snowflake{
			nodeID:   nodeID,
			sequence: 0,
			lastTime: 0,
		}
	})
	return err
}

func NextID() (int64, error) {
	if defaultNode == nil {
		return 0, errors.New("snowflake not initialized, call Init() first")
	}
	return defaultNode.Generate()
}

func New(nodeID int64) (*Snowflake, error) {
	if nodeID < 0 || nodeID > maxNodeID {
		return nil, ErrInvalidNodeID
	}
	return &Snowflake{
		nodeID:   nodeID,
		sequence: 0,
		lastTime: 0,
	}, nil
}

func (s *Snowflake) Generate() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli() - epoch

	if now < s.lastTime {
		return 0, errors.New("clock moved backwards")
	}

	if now == s.lastTime {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			for now <= s.lastTime {
				now = time.Now().UnixMilli() - epoch
			}
		}
	} else {
		s.sequence = 0
	}

	s.lastTime = now

	id := (now << timestampShift) | (s.nodeID << nodeIDShift) | s.sequence

	return id, nil
}

func Parse(id int64) (timestamp int64, nodeID int64, sequence int64) {
	timestamp = (id >> timestampShift) + epoch
	nodeID = (id >> nodeIDShift) & maxNodeID
	sequence = id & maxSequence
	return
}