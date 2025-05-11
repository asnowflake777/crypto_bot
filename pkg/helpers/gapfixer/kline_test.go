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
		name   string
		klines []*pgdb.Kline
		delta  int64
		want   []gap
	}{
		{
			name:   "no klines",
			klines: []*pgdb.Kline{},
			want:   nil,
		},
		{
			name: "one kline",
			klines: []*pgdb.Kline{
				{OpenTime: 0, CloseTime: 9},
			},
			want: nil,
		},
		{
			name:  "no gaps",
			delta: 1,
			klines: []*pgdb.Kline{
				{OpenTime: 0, CloseTime: 9},
				{OpenTime: 10, CloseTime: 19},
				{OpenTime: 20, CloseTime: 29},
				{OpenTime: 30, CloseTime: 39},
			},
			want: nil,
		},
		{
			name:  "one gap",
			delta: 10,
			klines: []*pgdb.Kline{
				{OpenTime: 0, CloseTime: 9},
				{OpenTime: 30, CloseTime: 39},
			},
			want: []gap{{start: 10, end: 30}},
		},
		{
			name:  "a few gaps",
			delta: 5,
			klines: []*pgdb.Kline{
				{OpenTime: 0, CloseTime: 9},
				{OpenTime: 30, CloseTime: 34},
				{OpenTime: 50, CloseTime: 59},
			},
			want: []gap{{start: 10, end: 30}, {start: 35, end: 50}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, findGaps(tc.klines, tc.delta))
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
		ReadKlines(gomock.Any(), pgdb.ReadKlinesRequest{Symbol: symbol, Interval: interval, OpenTime: from.UnixMilli(), CloseTime: to.UnixMilli(), Limit: 2}).
		Return([]*pgdb.Kline{
			{OpenTime: to.Add(-1 * time.Minute).UnixMilli(), CloseTime: to.Add(-1 * time.Millisecond).UnixMilli()},
		}, nil)

	ex.EXPECT().
		Klines(gomock.Any(), models.KlinesRequest{Symbol: symbol, Interval: interval, StartTime: from.UnixMilli(), EndTime: to.Add(-1*time.Minute - 1*time.Millisecond).UnixMilli(), Limit: 2}).
		Return([]*models.Kline{
			{OpenTime: from.UnixMilli(), CloseTime: from.Add(1*time.Minute - 1*time.Millisecond).UnixMilli()},
			{OpenTime: from.Add(1 * time.Minute).UnixMilli(), CloseTime: from.Add(2*time.Minute - 1*time.Millisecond).UnixMilli()},
		}, nil)
	s.EXPECT().WriteKlines(gomock.Any(), pgdb.WriteKlinesRequest{Symbol: symbol, Interval: interval,
		Klines: []*pgdb.Kline{
			{OpenTime: from.UnixMilli(), CloseTime: from.Add(1*time.Minute - 1*time.Millisecond).UnixMilli()},
			{OpenTime: from.Add(1 * time.Minute).UnixMilli(), CloseTime: from.Add(2*time.Minute - 1*time.Millisecond).UnixMilli()},
		},
	}).Return(nil, nil)

	ex.EXPECT().
		Klines(gomock.Any(), models.KlinesRequest{Symbol: symbol, Interval: interval, StartTime: from.Add(2 * time.Minute).UnixMilli(), EndTime: to.Add(-1*time.Minute - 1*time.Millisecond).UnixMilli(), Limit: 2}).
		Return([]*models.Kline{
			{OpenTime: from.Add(2 * time.Minute).UnixMilli(), CloseTime: from.Add(3*time.Minute - 1*time.Millisecond).UnixMilli()},
			{OpenTime: from.Add(3 * time.Minute).UnixMilli(), CloseTime: from.Add(4*time.Minute - 1*time.Millisecond).UnixMilli()},
		}, nil)
	s.EXPECT().WriteKlines(gomock.Any(), pgdb.WriteKlinesRequest{Symbol: symbol, Interval: interval,
		Klines: []*pgdb.Kline{
			{OpenTime: from.Add(2 * time.Minute).UnixMilli(), CloseTime: from.Add(3*time.Minute - 1*time.Millisecond).UnixMilli()},
			{OpenTime: from.Add(3 * time.Minute).UnixMilli(), CloseTime: from.Add(4*time.Minute - 1*time.Millisecond).UnixMilli()},
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
		ReadKlines(gomock.Any(), pgdb.ReadKlinesRequest{Symbol: symbol, Interval: interval, OpenTime: from.UnixMilli(), CloseTime: to.UnixMilli(), Limit: 2}).
		Return([]*pgdb.Kline{
			{OpenTime: from.UnixMilli(), CloseTime: from.Add(1*time.Minute - 1*time.Millisecond).UnixMilli()},
		}, nil)

	ex.EXPECT().
		Klines(gomock.Any(), models.KlinesRequest{Symbol: symbol, Interval: interval, StartTime: from.Add(1 * time.Minute).UnixMilli(), EndTime: to.UnixMilli(), Limit: 2}).
		Return([]*models.Kline{
			{OpenTime: from.Add(1 * time.Minute).UnixMilli(), CloseTime: from.Add(2*time.Minute - 1*time.Millisecond).UnixMilli()},
			{OpenTime: from.Add(2 * time.Minute).UnixMilli(), CloseTime: from.Add(3*time.Minute - 1*time.Millisecond).UnixMilli()},
		}, nil)
	s.EXPECT().WriteKlines(gomock.Any(), pgdb.WriteKlinesRequest{Symbol: symbol, Interval: interval,
		Klines: []*pgdb.Kline{
			{OpenTime: from.Add(1 * time.Minute).UnixMilli(), CloseTime: from.Add(2*time.Minute - 1*time.Millisecond).UnixMilli()},
			{OpenTime: from.Add(2 * time.Minute).UnixMilli(), CloseTime: from.Add(3*time.Minute - 1*time.Millisecond).UnixMilli()},
		},
	}).Return(nil, nil)

	ex.EXPECT().
		Klines(gomock.Any(), models.KlinesRequest{Symbol: symbol, Interval: interval, StartTime: from.Add(3 * time.Minute).UnixMilli(), EndTime: to.UnixMilli(), Limit: 2}).
		Return([]*models.Kline{
			{OpenTime: from.Add(3 * time.Minute).UnixMilli(), CloseTime: from.Add(4*time.Minute - 1*time.Millisecond).UnixMilli()},
			{OpenTime: from.Add(4 * time.Minute).UnixMilli(), CloseTime: to.Add(5*time.Minute - 1*time.Millisecond).UnixMilli()},
		}, nil)
	s.EXPECT().WriteKlines(gomock.Any(), pgdb.WriteKlinesRequest{Symbol: symbol, Interval: interval,
		Klines: []*pgdb.Kline{
			{OpenTime: from.Add(3 * time.Minute).UnixMilli(), CloseTime: from.Add(4*time.Minute - 1*time.Millisecond).UnixMilli()},
			{OpenTime: from.Add(4 * time.Minute).UnixMilli(), CloseTime: to.Add(5*time.Minute - 1*time.Millisecond).UnixMilli()},
		},
	}).Return(nil, nil)

	err := FixGaps(context.Background(), ex, s, symbol, interval, from, to, 2)
	require.NoError(t, err)
}
