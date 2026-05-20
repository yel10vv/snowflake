package snowflake

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestNewWithoutOpts(t *testing.T) {
	var idMachine uint16 = 1
	var start = time.Now().UnixMilli()

	sf, err := New(idMachine)
	assert.Equal(t, nil, err, "Произошла ошибка")
	assert.Equal(t, idMachine, sf.idMachine, "Разные id")
	assert.Equal(t,  start, sf.startTIme,"Разное время")

}

func TestNewWithTime(t *testing.T) {
	var idMachine uint16 = 1
	testcases := []struct {
		name  string
		start time.Time
		err   error
	}{
		{
			name:  "Bad time",
			start: time.Now().Add(time.Hour),
			err:   ErrStartTimeAhead,
		},
		{
			name:  "Good time",
			start: time.Now(),
			err:   nil,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			sf, err := New(idMachine, WithStartTime(tc.start))
			require.ErrorIs(t, tc.err, err)
			if err != nil {
				return
			}
			assert.Equal(t, idMachine, sf.idMachine)
			assert.Equal(t, tc.start.UnixMilli(), sf.startTIme)
		})
	}
}

func TestNextID(t *testing.T) {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	iter := 2
	sf, err := New(1)
	require.NoError(t, err)

	num := 8000
	set := make(map[uint64]struct{}, num*2)
	var g errgroup.Group
	var mx sync.Mutex
	for j := 0; j < iter; j++ {
		for i := 0; i < num; i++ {
			g.Go(func() error {
				ID, err := sf.NextID()
				if err != nil {
					return err
				}
				mx.Lock()
				set[ID] = struct{}{}
				mx.Unlock()
				return nil
			})
		}
		time.Sleep(period)
	}

	if err := g.Wait(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assert.Equal(t, num*iter, len(set))

}

func TestNextIDError(t *testing.T) {
	sf, err := New(1)
	require.NoError(t, err)

	sf.elapsedTime = 1 << BitLenTime
	_, err = sf.NextID()
	require.ErrorIs(t, ErrOverTimeLimit, err)
}


func TestDocomposite(t *testing.T){
	var(
		iter = 5
		expectedSeq uint64 = 1
		expectedMachine uint16 = 42
		expectedTime uint64 = 0
	)
	sf, err := New(expectedMachine)
	if err != nil {
		t.Fatalf("init error: %v", err)
	}
			
	for i := 0; i < iter; i++{		
		id, err := sf.NextID()
		if err != nil {
			t.Fatalf("critical error: %v", err)
		}
		seq := 0xFFF & id
		machine := uint16((id >> BitLenSequence) & 0x3FF)
		time := (id >> (BitLenSequence + BitLenMachineID)) & 0x1FFFFFFFFFF 
		assert.Equal(t, expectedSeq, seq)
		assert.Equal(t, expectedMachine, machine)
		assert.Equal(t, expectedTime, time)
		expectedSeq++
	}

}