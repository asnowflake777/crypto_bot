package gapfixer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"crypto_bot/pkg/exchange/models"
	mockgapfixer "crypto_bot/pkg/helpers/gapfixer/mocks"
	"crypto_bot/pkg/storage/pgdb"
)

func TestFindGaps(t *testing.T) {
	testCases := []struct {
		name      string
		klines    []*pgdb.Kline
		from, to  int64
		chunkSize int
		want      []gap
	}{
		{
			name:   "no params",
			klines: []*pgdb.Kline{},
			want:   []gap{{start: 0, end: 0}},
		},
		{
			name: "one kline",
			klines: []*pgdb.Kline{
				{OpenTime: 0, CloseTime: 9},
			},
			from:      0,
			to:        10,
			chunkSize: 1,
			want:      nil,
		},
		{
			name: "no gaps",
			klines: []*pgdb.Kline{
				{OpenTime: 0, CloseTime: 9},
				{OpenTime: 10, CloseTime: 19},
				{OpenTime: 20, CloseTime: 29},
				{OpenTime: 30, CloseTime: 39},
			},
			from:      0,
			to:        40,
			chunkSize: 4,
			want:      nil,
		},
		{
			name: "gap in middle",
			klines: []*pgdb.Kline{
				{OpenTime: 0, CloseTime: 9},
				{OpenTime: 30, CloseTime: 39},
				{OpenTime: 40, CloseTime: 49},
				{OpenTime: 50, CloseTime: 59},
			},
			from:      0,
			to:        60,
			chunkSize: 4,
			want:      []gap{{start: 10, end: 29}},
		},
		{
			name: "a few gaps in middle",
			klines: []*pgdb.Kline{
				{OpenTime: 0, CloseTime: 9},
				{OpenTime: 30, CloseTime: 39},
				{OpenTime: 50, CloseTime: 59},
			},
			from:      0,
			chunkSize: 6,
			want:      []gap{{start: 10, end: 29}, {start: 40, end: 49}},
		},
		{
			name: "all possible gaps",
			klines: []*pgdb.Kline{
				{OpenTime: 10, CloseTime: 19},
				{OpenTime: 30, CloseTime: 39},
				{OpenTime: 50, CloseTime: 59},
			},
			from:      0,
			to:        100,
			chunkSize: 3,
			want:      []gap{{start: 0, end: 9}, {start: 20, end: 29}, {start: 40, end: 49}},
		},
		{
			name: "empty start",
			klines: []*pgdb.Kline{
				{OpenTime: 40, CloseTime: 49},
				{OpenTime: 50, CloseTime: 59},
			},
			from:      0,
			chunkSize: 6,
			want:      []gap{{start: 0, end: 39}},
		},
		{
			name: "empty end",
			klines: []*pgdb.Kline{
				{OpenTime: 30, CloseTime: 39},
				{OpenTime: 40, CloseTime: 49},
				{OpenTime: 50, CloseTime: 59},
			},
			from:      30,
			to:        90,
			chunkSize: 6,
			want:      []gap{{start: 60, end: 90}},
		},
		{
			name:   "empty klines",
			klines: []*pgdb.Kline{},
			from:   0,
			to:     60,
			want:   []gap{{start: 0, end: 60}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, findGaps(tc.klines, tc.from, tc.to, tc.chunkSize))
		})
	}
}

func TestFixGaps_NoStartKlines(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := mockgapfixer.NewMockStorage(ctrl)
	ex := mockgapfixer.NewMockExchange(ctrl)

	symbol, interval := "symbol", "1m"
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(5 * time.Minute)

	s.EXPECT().
		ReadKlines(gomock.Any(), pgdb.ReadKlinesRequest{Symbol: symbol, Interval: interval,
			OpenTime:  from.UnixMilli(),
			CloseTime: to.UnixMilli(), Limit: 2}).
		Return([]*pgdb.Kline{
			{OpenTime: to.Add(-1 * time.Minute).UnixMilli(), CloseTime: to.Add(-1 * time.Millisecond).UnixMilli()},
		}, nil)

	ex.EXPECT().
		Klines(gomock.Any(), models.KlinesRequest{Symbol: symbol, Interval: interval,
			StartTime: from.UnixMilli(),
			EndTime:   to.Add(-1*time.Minute).UnixMilli() - 1, Limit: 2}).
		Return([]*models.Kline{
			{OpenTime: from.UnixMilli(), CloseTime: from.Add(1*time.Minute).UnixMilli() - 1},
			{OpenTime: from.Add(1 * time.Minute).UnixMilli(), CloseTime: from.Add(2*time.Minute).UnixMilli() - 1},
		}, nil)
	s.EXPECT().WriteKlines(gomock.Any(), pgdb.WriteKlinesRequest{Symbol: symbol, Interval: interval,
		Klines: []*pgdb.Kline{
			{OpenTime: from.UnixMilli(), CloseTime: from.Add(1*time.Minute).UnixMilli() - 1},
			{OpenTime: from.Add(1 * time.Minute).UnixMilli(), CloseTime: from.Add(2*time.Minute).UnixMilli() - 1},
		},
	}).Return(nil, nil)

	ex.EXPECT().
		Klines(gomock.Any(), models.KlinesRequest{Symbol: symbol, Interval: interval,
			StartTime: from.Add(2 * time.Minute).UnixMilli(),
			EndTime:   to.Add(-1*time.Minute).UnixMilli() - 1, Limit: 2}).
		Return([]*models.Kline{
			{OpenTime: from.Add(2 * time.Minute).UnixMilli(), CloseTime: from.Add(3*time.Minute).UnixMilli() - 1},
			{OpenTime: from.Add(3 * time.Minute).UnixMilli(), CloseTime: from.Add(4*time.Minute).UnixMilli() - 1},
		}, nil)
	s.EXPECT().WriteKlines(gomock.Any(), pgdb.WriteKlinesRequest{Symbol: symbol, Interval: interval,
		Klines: []*pgdb.Kline{
			{OpenTime: from.Add(2 * time.Minute).UnixMilli(), CloseTime: from.Add(3*time.Minute).UnixMilli() - 1},
			{OpenTime: from.Add(3 * time.Minute).UnixMilli(), CloseTime: from.Add(4*time.Minute).UnixMilli() - 1},
		},
	}).Return(nil, nil)

	err := FixGaps(context.Background(), ex, s, symbol, interval, from, to, 2)
	require.NoError(t, err)
}

func TestFixGaps_NoEndKlines(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := mockgapfixer.NewMockStorage(ctrl)
	ex := mockgapfixer.NewMockExchange(ctrl)

	symbol, interval := "symbol", "1m"
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(5 * time.Minute)

	s.EXPECT().
		ReadKlines(gomock.Any(), pgdb.ReadKlinesRequest{Symbol: symbol, Interval: interval,
			OpenTime:  from.UnixMilli(),
			CloseTime: to.UnixMilli(), Limit: 2}).
		Return([]*pgdb.Kline{
			{OpenTime: from.UnixMilli(), CloseTime: from.Add(1*time.Minute).UnixMilli() - 1},
		}, nil)

	ex.EXPECT().
		Klines(gomock.Any(), models.KlinesRequest{Symbol: symbol, Interval: interval,
			StartTime: from.Add(1 * time.Minute).UnixMilli(),
			EndTime:   to.UnixMilli(), Limit: 2}).
		Return([]*models.Kline{
			{OpenTime: from.Add(1 * time.Minute).UnixMilli(), CloseTime: from.Add(2*time.Minute).UnixMilli() - 1},
			{OpenTime: from.Add(2 * time.Minute).UnixMilli(), CloseTime: from.Add(3*time.Minute).UnixMilli() - 1},
		}, nil)
	s.EXPECT().WriteKlines(gomock.Any(), pgdb.WriteKlinesRequest{Symbol: symbol, Interval: interval,
		Klines: []*pgdb.Kline{
			{OpenTime: from.Add(1 * time.Minute).UnixMilli(), CloseTime: from.Add(2*time.Minute).UnixMilli() - 1},
			{OpenTime: from.Add(2 * time.Minute).UnixMilli(), CloseTime: from.Add(3*time.Minute).UnixMilli() - 1},
		},
	}).Return(nil, nil)

	ex.EXPECT().
		Klines(gomock.Any(), models.KlinesRequest{Symbol: symbol, Interval: interval,
			StartTime: from.Add(3 * time.Minute).UnixMilli(),
			EndTime:   to.UnixMilli(), Limit: 2}).
		Return([]*models.Kline{
			{OpenTime: from.Add(3 * time.Minute).UnixMilli(), CloseTime: from.Add(4*time.Minute).UnixMilli() - 1},
			{OpenTime: from.Add(4 * time.Minute).UnixMilli(), CloseTime: to.Add(5*time.Minute).UnixMilli() - 1},
		}, nil)
	s.EXPECT().WriteKlines(gomock.Any(), pgdb.WriteKlinesRequest{Symbol: symbol, Interval: interval,
		Klines: []*pgdb.Kline{
			{OpenTime: from.Add(3 * time.Minute).UnixMilli(), CloseTime: from.Add(4*time.Minute).UnixMilli() - 1},
			{OpenTime: from.Add(4 * time.Minute).UnixMilli(), CloseTime: to.Add(5*time.Minute).UnixMilli() - 1},
		},
	}).Return(nil, nil)

	err := FixGaps(context.Background(), ex, s, symbol, interval, from, to, 2)
	require.NoError(t, err)
}
