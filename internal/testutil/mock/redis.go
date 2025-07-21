package mock

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
)

type RedisClient struct {
	mock.Mock
}

func (r *RedisClient) TFunctionLoad(ctx context.Context, lib string) *redis.StatusCmd {
	args := r.Called(ctx, lib)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) TFunctionLoadArgs(ctx context.Context, lib string, options *redis.TFunctionLoadOptions) *redis.StatusCmd {
	args := r.Called(ctx, lib, options)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) TFunctionDelete(ctx context.Context, libName string) *redis.StatusCmd {
	args := r.Called(ctx, libName)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) TFunctionList(ctx context.Context) *redis.MapStringInterfaceSliceCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.MapStringInterfaceSliceCmd)
}

func (r *RedisClient) TFunctionListArgs(ctx context.Context, options *redis.TFunctionListOptions) *redis.MapStringInterfaceSliceCmd {
	args := r.Called(ctx, options)
	return args.Get(0).(*redis.MapStringInterfaceSliceCmd)
}

func (r *RedisClient) TFCall(ctx context.Context, libName string, funcName string, numKeys int) *redis.Cmd {
	args := r.Called(ctx, libName, funcName, numKeys)
	return args.Get(0).(*redis.Cmd)
}

func (r *RedisClient) TFCallArgs(ctx context.Context, libName string, funcName string, numKeys int, options *redis.TFCallOptions) *redis.Cmd {
	args := r.Called(ctx, libName, funcName, numKeys, options)
	return args.Get(0).(*redis.Cmd)
}

func (r *RedisClient) TFCallASYNC(ctx context.Context, libName string, funcName string, numKeys int) *redis.Cmd {
	args := r.Called(ctx, libName, funcName, numKeys)
	return args.Get(0).(*redis.Cmd)
}

func (r *RedisClient) TFCallASYNCArgs(ctx context.Context, libName string, funcName string, numKeys int, options *redis.TFCallOptions) *redis.Cmd {
	args := r.Called(ctx, libName, funcName, numKeys, options)
	return args.Get(0).(*redis.Cmd)
}

func (r *RedisClient) ObjectFreq(ctx context.Context, key string) *redis.IntCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) BFAdd(ctx context.Context, key string, element interface{}) *redis.BoolCmd {
	args := r.Called(ctx, key, element)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) BFCard(ctx context.Context, key string) *redis.IntCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) BFExists(ctx context.Context, key string, element interface{}) *redis.BoolCmd {
	args := r.Called(ctx, key, element)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) BFInfo(ctx context.Context, key string) *redis.BFInfoCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.BFInfoCmd)
}

func (r *RedisClient) BFInfoArg(ctx context.Context, key, option string) *redis.BFInfoCmd {
	args := r.Called(ctx, key, option)
	return args.Get(0).(*redis.BFInfoCmd)
}

func (r *RedisClient) BFInfoCapacity(ctx context.Context, key string) *redis.BFInfoCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.BFInfoCmd)
}

func (r *RedisClient) BFInfoSize(ctx context.Context, key string) *redis.BFInfoCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.BFInfoCmd)
}

func (r *RedisClient) BFInfoFilters(ctx context.Context, key string) *redis.BFInfoCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.BFInfoCmd)
}

func (r *RedisClient) BFInfoItems(ctx context.Context, key string) *redis.BFInfoCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.BFInfoCmd)
}

func (r *RedisClient) BFInfoExpansion(ctx context.Context, key string) *redis.BFInfoCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.BFInfoCmd)
}

func (r *RedisClient) BFInsert(ctx context.Context, key string, options *redis.BFInsertOptions, elements ...interface{}) *redis.BoolSliceCmd {
	args := r.Called(ctx, key, options, elements)
	return args.Get(0).(*redis.BoolSliceCmd)
}

func (r *RedisClient) BFMAdd(ctx context.Context, key string, elements ...interface{}) *redis.BoolSliceCmd {
	args := r.Called(ctx, key, elements)
	return args.Get(0).(*redis.BoolSliceCmd)
}

func (r *RedisClient) BFMExists(ctx context.Context, key string, elements ...interface{}) *redis.BoolSliceCmd {
	args := r.Called(ctx, key, elements)
	return args.Get(0).(*redis.BoolSliceCmd)
}

func (r *RedisClient) BFReserve(ctx context.Context, key string, errorRate float64, capacity int64) *redis.StatusCmd {
	args := r.Called(ctx, key, errorRate, capacity)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) BFReserveExpansion(ctx context.Context, key string, errorRate float64, capacity, expansion int64) *redis.StatusCmd {
	args := r.Called(ctx, key, errorRate, capacity, expansion)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) BFReserveNonScaling(ctx context.Context, key string, errorRate float64, capacity int64) *redis.StatusCmd {
	args := r.Called(ctx, key, errorRate, capacity)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) BFReserveWithArgs(ctx context.Context, key string, options *redis.BFReserveOptions) *redis.StatusCmd {
	args := r.Called(ctx, key, options)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) BFScanDump(ctx context.Context, key string, iterator int64) *redis.ScanDumpCmd {
	args := r.Called(ctx, key, iterator)
	return args.Get(0).(*redis.ScanDumpCmd)
}

func (r *RedisClient) BFLoadChunk(ctx context.Context, key string, iterator int64, data interface{}) *redis.StatusCmd {
	args := r.Called(ctx, key, iterator, data)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) CFAdd(ctx context.Context, key string, element interface{}) *redis.BoolCmd {
	args := r.Called(ctx, key, element)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) CFAddNX(ctx context.Context, key string, element interface{}) *redis.BoolCmd {
	args := r.Called(ctx, key, element)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) CFCount(ctx context.Context, key string, element interface{}) *redis.IntCmd {
	args := r.Called(ctx, key, element)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) CFDel(ctx context.Context, key string, element interface{}) *redis.BoolCmd {
	args := r.Called(ctx, key, element)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) CFExists(ctx context.Context, key string, element interface{}) *redis.BoolCmd {
	args := r.Called(ctx, key, element)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) CFInfo(ctx context.Context, key string) *redis.CFInfoCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.CFInfoCmd)
}

func (r *RedisClient) CFInsert(ctx context.Context, key string, options *redis.CFInsertOptions, elements ...interface{}) *redis.BoolSliceCmd {
	args := r.Called(ctx, key, options, elements)
	return args.Get(0).(*redis.BoolSliceCmd)
}

func (r *RedisClient) CFInsertNX(ctx context.Context, key string, options *redis.CFInsertOptions, elements ...interface{}) *redis.IntSliceCmd {
	args := r.Called(ctx, key, options, elements)
	return args.Get(0).(*redis.IntSliceCmd)
}

func (r *RedisClient) CFMExists(ctx context.Context, key string, elements ...interface{}) *redis.BoolSliceCmd {
	args := r.Called(ctx, key, elements)
	return args.Get(0).(*redis.BoolSliceCmd)
}

func (r *RedisClient) CFReserve(ctx context.Context, key string, capacity int64) *redis.StatusCmd {
	args := r.Called(ctx, key, capacity)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) CFReserveWithArgs(ctx context.Context, key string, options *redis.CFReserveOptions) *redis.StatusCmd {
	args := r.Called(ctx, key, options)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) CFReserveExpansion(ctx context.Context, key string, capacity int64, expansion int64) *redis.StatusCmd {
	args := r.Called(ctx, key, capacity, expansion)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) CFReserveBucketSize(ctx context.Context, key string, capacity int64, bucketsize int64) *redis.StatusCmd {
	args := r.Called(ctx, key, capacity, bucketsize)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) CFReserveMaxIterations(ctx context.Context, key string, capacity int64, maxiterations int64) *redis.StatusCmd {
	args := r.Called(ctx, key, capacity, maxiterations)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) CFScanDump(ctx context.Context, key string, iterator int64) *redis.ScanDumpCmd {
	args := r.Called(ctx, key, iterator)
	return args.Get(0).(*redis.ScanDumpCmd)
}

func (r *RedisClient) CFLoadChunk(ctx context.Context, key string, iterator int64, data interface{}) *redis.StatusCmd {
	args := r.Called(ctx, key, iterator, data)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) CMSIncrBy(ctx context.Context, key string, elements ...interface{}) *redis.IntSliceCmd {
	args := r.Called(ctx, key, elements)
	return args.Get(0).(*redis.IntSliceCmd)
}

func (r *RedisClient) CMSInfo(ctx context.Context, key string) *redis.CMSInfoCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.CMSInfoCmd)
}

func (r *RedisClient) CMSInitByDim(ctx context.Context, key string, width, height int64) *redis.StatusCmd {
	args := r.Called(ctx, key, width, height)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) CMSInitByProb(ctx context.Context, key string, errorRate, probability float64) *redis.StatusCmd {
	args := r.Called(ctx, key, errorRate, probability)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) CMSMerge(ctx context.Context, destKey string, sourceKeys ...string) *redis.StatusCmd {
	args := r.Called(ctx, destKey, sourceKeys)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) CMSMergeWithWeight(ctx context.Context, destKey string, sourceKeys map[string]int64) *redis.StatusCmd {
	args := r.Called(ctx, destKey, sourceKeys)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) CMSQuery(ctx context.Context, key string, elements ...interface{}) *redis.IntSliceCmd {
	args := r.Called(ctx, key, elements)
	return args.Get(0).(*redis.IntSliceCmd)
}

func (r *RedisClient) TopKAdd(ctx context.Context, key string, elements ...interface{}) *redis.StringSliceCmd {
	args := r.Called(ctx, key, elements)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) TopKCount(ctx context.Context, key string, elements ...interface{}) *redis.IntSliceCmd {
	args := r.Called(ctx, key, elements)
	return args.Get(0).(*redis.IntSliceCmd)
}

func (r *RedisClient) TopKIncrBy(ctx context.Context, key string, elements ...interface{}) *redis.StringSliceCmd {
	args := r.Called(ctx, key, elements)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) TopKInfo(ctx context.Context, key string) *redis.TopKInfoCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.TopKInfoCmd)
}

func (r *RedisClient) TopKList(ctx context.Context, key string) *redis.StringSliceCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) TopKListWithCount(ctx context.Context, key string) *redis.MapStringIntCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.MapStringIntCmd)
}

func (r *RedisClient) TopKQuery(ctx context.Context, key string, elements ...interface{}) *redis.BoolSliceCmd {
	args := r.Called(ctx, key, elements)
	return args.Get(0).(*redis.BoolSliceCmd)
}

func (r *RedisClient) TopKReserve(ctx context.Context, key string, k int64) *redis.StatusCmd {
	args := r.Called(ctx, key, k)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) TopKReserveWithOptions(ctx context.Context, key string, k int64, width, depth int64, decay float64) *redis.StatusCmd {
	args := r.Called(ctx, key, k, width, depth, decay)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) TDigestAdd(ctx context.Context, key string, elements ...float64) *redis.StatusCmd {
	args := r.Called(ctx, key, elements)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) TDigestByRank(ctx context.Context, key string, rank ...uint64) *redis.FloatSliceCmd {
	args := r.Called(ctx, key, rank)
	return args.Get(0).(*redis.FloatSliceCmd)
}

func (r *RedisClient) TDigestByRevRank(ctx context.Context, key string, rank ...uint64) *redis.FloatSliceCmd {
	args := r.Called(ctx, key, rank)
	return args.Get(0).(*redis.FloatSliceCmd)
}

func (r *RedisClient) TDigestCDF(ctx context.Context, key string, elements ...float64) *redis.FloatSliceCmd {
	args := r.Called(ctx, key, elements)
	return args.Get(0).(*redis.FloatSliceCmd)
}

func (r *RedisClient) TDigestCreate(ctx context.Context, key string) *redis.StatusCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) TDigestCreateWithCompression(ctx context.Context, key string, compression int64) *redis.StatusCmd {
	args := r.Called(ctx, key, compression)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) TDigestInfo(ctx context.Context, key string) *redis.TDigestInfoCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.TDigestInfoCmd)
}

func (r *RedisClient) TDigestMax(ctx context.Context, key string) *redis.FloatCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.FloatCmd)
}

func (r *RedisClient) TDigestMin(ctx context.Context, key string) *redis.FloatCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.FloatCmd)
}

func (r *RedisClient) TDigestMerge(ctx context.Context, destKey string, options *redis.TDigestMergeOptions, sourceKeys ...string) *redis.StatusCmd {
	args := r.Called(ctx, destKey, options, sourceKeys)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) TDigestQuantile(ctx context.Context, key string, elements ...float64) *redis.FloatSliceCmd {
	args := r.Called(ctx, key, elements)
	return args.Get(0).(*redis.FloatSliceCmd)
}

func (r *RedisClient) TDigestRank(ctx context.Context, key string, values ...float64) *redis.IntSliceCmd {
	args := r.Called(ctx, key, values)
	return args.Get(0).(*redis.IntSliceCmd)
}

func (r *RedisClient) TDigestReset(ctx context.Context, key string) *redis.StatusCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) TDigestRevRank(ctx context.Context, key string, values ...float64) *redis.IntSliceCmd {
	args := r.Called(ctx, key, values)
	return args.Get(0).(*redis.IntSliceCmd)
}

func (r *RedisClient) TDigestTrimmedMean(ctx context.Context, key string, lowCutQuantile, highCutQuantile float64) *redis.FloatCmd {
	args := r.Called(ctx, key, lowCutQuantile, highCutQuantile)
	return args.Get(0).(*redis.FloatCmd)
}

func (r *RedisClient) TSAdd(ctx context.Context, key string, timestamp interface{}, value float64) *redis.IntCmd {
	args := r.Called(ctx, key, timestamp, value)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) TSAddWithArgs(ctx context.Context, key string, timestamp interface{}, value float64, options *redis.TSOptions) *redis.IntCmd {
	args := r.Called(ctx, key, timestamp, value, options)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) TSCreate(ctx context.Context, key string) *redis.StatusCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) TSCreateWithArgs(ctx context.Context, key string, options *redis.TSOptions) *redis.StatusCmd {
	args := r.Called(ctx, key, options)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) TSAlter(ctx context.Context, key string, options *redis.TSAlterOptions) *redis.StatusCmd {
	args := r.Called(ctx, key, options)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) TSCreateRule(ctx context.Context, sourceKey string, destKey string, aggregator redis.Aggregator, bucketDuration int) *redis.StatusCmd {
	args := r.Called(ctx, sourceKey, destKey, aggregator, bucketDuration)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) TSCreateRuleWithArgs(ctx context.Context, sourceKey string, destKey string, aggregator redis.Aggregator, bucketDuration int, options *redis.TSCreateRuleOptions) *redis.StatusCmd {
	args := r.Called(ctx, sourceKey, destKey, aggregator, bucketDuration, options)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) TSIncrBy(ctx context.Context, key string, timestamp float64) *redis.IntCmd {
	args := r.Called(ctx, key, timestamp)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) TSIncrByWithArgs(ctx context.Context, key string, timestamp float64, options *redis.TSIncrDecrOptions) *redis.IntCmd {
	args := r.Called(ctx, key, timestamp, options)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) TSDecrBy(ctx context.Context, key string, timestamp float64) *redis.IntCmd {
	args := r.Called(ctx, key, timestamp)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) TSDecrByWithArgs(ctx context.Context, key string, timestamp float64, options *redis.TSIncrDecrOptions) *redis.IntCmd {
	args := r.Called(ctx, key, timestamp, options)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) TSDel(ctx context.Context, key string, fromTimestamp int, toTimestamp int) *redis.IntCmd {
	args := r.Called(ctx, key, fromTimestamp, toTimestamp)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) TSDeleteRule(ctx context.Context, sourceKey string, destKey string) *redis.StatusCmd {
	args := r.Called(ctx, sourceKey, destKey)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) TSGet(ctx context.Context, key string) *redis.TSTimestampValueCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.TSTimestampValueCmd)
}

func (r *RedisClient) TSGetWithArgs(ctx context.Context, key string, options *redis.TSGetOptions) *redis.TSTimestampValueCmd {
	args := r.Called(ctx, key, options)
	return args.Get(0).(*redis.TSTimestampValueCmd)
}

func (r *RedisClient) TSInfo(ctx context.Context, key string) *redis.MapStringInterfaceCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.MapStringInterfaceCmd)
}

func (r *RedisClient) TSInfoWithArgs(ctx context.Context, key string, options *redis.TSInfoOptions) *redis.MapStringInterfaceCmd {
	args := r.Called(ctx, key, options)
	return args.Get(0).(*redis.MapStringInterfaceCmd)
}

func (r *RedisClient) TSMAdd(ctx context.Context, ktvSlices [][]interface{}) *redis.IntSliceCmd {
	args := r.Called(ctx, ktvSlices)
	return args.Get(0).(*redis.IntSliceCmd)
}

func (r *RedisClient) TSQueryIndex(ctx context.Context, filterExpr []string) *redis.StringSliceCmd {
	args := r.Called(ctx, filterExpr)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) TSRevRange(ctx context.Context, key string, fromTimestamp int, toTimestamp int) *redis.TSTimestampValueSliceCmd {
	args := r.Called(ctx, key, fromTimestamp, toTimestamp)
	return args.Get(0).(*redis.TSTimestampValueSliceCmd)
}

func (r *RedisClient) TSRevRangeWithArgs(ctx context.Context, key string, fromTimestamp int, toTimestamp int, options *redis.TSRevRangeOptions) *redis.TSTimestampValueSliceCmd {
	args := r.Called(ctx, key, fromTimestamp, toTimestamp, options)
	return args.Get(0).(*redis.TSTimestampValueSliceCmd)
}

func (r *RedisClient) TSRange(ctx context.Context, key string, fromTimestamp int, toTimestamp int) *redis.TSTimestampValueSliceCmd {
	args := r.Called(ctx, key, fromTimestamp, toTimestamp)
	return args.Get(0).(*redis.TSTimestampValueSliceCmd)
}

func (r *RedisClient) TSRangeWithArgs(ctx context.Context, key string, fromTimestamp int, toTimestamp int, options *redis.TSRangeOptions) *redis.TSTimestampValueSliceCmd {
	args := r.Called(ctx, key, fromTimestamp, toTimestamp, options)
	return args.Get(0).(*redis.TSTimestampValueSliceCmd)
}

func (r *RedisClient) TSMRange(ctx context.Context, fromTimestamp int, toTimestamp int, filterExpr []string) *redis.MapStringSliceInterfaceCmd {
	args := r.Called(ctx, fromTimestamp, toTimestamp, filterExpr)
	return args.Get(0).(*redis.MapStringSliceInterfaceCmd)
}

func (r *RedisClient) TSMRangeWithArgs(ctx context.Context, fromTimestamp int, toTimestamp int, filterExpr []string, options *redis.TSMRangeOptions) *redis.MapStringSliceInterfaceCmd {
	args := r.Called(ctx, fromTimestamp, toTimestamp, filterExpr, options)
	return args.Get(0).(*redis.MapStringSliceInterfaceCmd)
}

func (r *RedisClient) TSMRevRange(ctx context.Context, fromTimestamp int, toTimestamp int, filterExpr []string) *redis.MapStringSliceInterfaceCmd {
	args := r.Called(ctx, fromTimestamp, toTimestamp, filterExpr)
	return args.Get(0).(*redis.MapStringSliceInterfaceCmd)
}

func (r *RedisClient) TSMRevRangeWithArgs(ctx context.Context, fromTimestamp int, toTimestamp int, filterExpr []string, options *redis.TSMRevRangeOptions) *redis.MapStringSliceInterfaceCmd {
	args := r.Called(ctx, fromTimestamp, toTimestamp, filterExpr, options)
	return args.Get(0).(*redis.MapStringSliceInterfaceCmd)
}

func (r *RedisClient) TSMGet(ctx context.Context, filters []string) *redis.MapStringSliceInterfaceCmd {
	args := r.Called(ctx, filters)
	return args.Get(0).(*redis.MapStringSliceInterfaceCmd)
}

func (r *RedisClient) TSMGetWithArgs(ctx context.Context, filters []string, options *redis.TSMGetOptions) *redis.MapStringSliceInterfaceCmd {
	args := r.Called(ctx, filters, options)
	return args.Get(0).(*redis.MapStringSliceInterfaceCmd)
}

func (r *RedisClient) JSONArrAppend(ctx context.Context, key, path string, values ...interface{}) *redis.IntSliceCmd {
	args := r.Called(ctx, key, path, values)
	return args.Get(0).(*redis.IntSliceCmd)
}

func (r *RedisClient) JSONArrIndex(ctx context.Context, key, path string, value ...interface{}) *redis.IntSliceCmd {
	args := r.Called(ctx, key, path, value)
	return args.Get(0).(*redis.IntSliceCmd)
}

func (r *RedisClient) JSONArrIndexWithArgs(ctx context.Context, key, path string, options *redis.JSONArrIndexArgs, value ...interface{}) *redis.IntSliceCmd {
	args := r.Called(ctx, key, path, options, value)
	return args.Get(0).(*redis.IntSliceCmd)
}

func (r *RedisClient) JSONArrInsert(ctx context.Context, key, path string, index int64, values ...interface{}) *redis.IntSliceCmd {
	args := r.Called(ctx, key, path, index, values)
	return args.Get(0).(*redis.IntSliceCmd)
}

func (r *RedisClient) JSONArrLen(ctx context.Context, key, path string) *redis.IntSliceCmd {
	args := r.Called(ctx, key, path)
	return args.Get(0).(*redis.IntSliceCmd)
}

func (r *RedisClient) JSONArrPop(ctx context.Context, key, path string, index int) *redis.StringSliceCmd {
	args := r.Called(ctx, key, path, index)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) JSONArrTrim(ctx context.Context, key, path string) *redis.IntSliceCmd {
	args := r.Called(ctx, key, path)
	return args.Get(0).(*redis.IntSliceCmd)
}

func (r *RedisClient) JSONArrTrimWithArgs(ctx context.Context, key, path string, options *redis.JSONArrTrimArgs) *redis.IntSliceCmd {
	args := r.Called(ctx, key, path, options)
	return args.Get(0).(*redis.IntSliceCmd)
}

func (r *RedisClient) JSONClear(ctx context.Context, key, path string) *redis.IntCmd {
	args := r.Called(ctx, key, path)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) JSONDebugMemory(ctx context.Context, key, path string) *redis.IntCmd {
	args := r.Called(ctx, key, path)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) JSONDel(ctx context.Context, key, path string) *redis.IntCmd {
	args := r.Called(ctx, key, path)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) JSONForget(ctx context.Context, key, path string) *redis.IntCmd {
	args := r.Called(ctx, key, path)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) JSONGet(ctx context.Context, key string, paths ...string) *redis.JSONCmd {
	args := r.Called(ctx, key, paths)
	return args.Get(0).(*redis.JSONCmd)
}

func (r *RedisClient) JSONGetWithArgs(ctx context.Context, key string, options *redis.JSONGetArgs, paths ...string) *redis.JSONCmd {
	args := r.Called(ctx, key, options, paths)
	return args.Get(0).(*redis.JSONCmd)
}

func (r *RedisClient) JSONMerge(ctx context.Context, key, path string, value string) *redis.StatusCmd {
	args := r.Called(ctx, key, path, value)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) JSONMSetArgs(ctx context.Context, docs []redis.JSONSetArgs) *redis.StatusCmd {
	args := r.Called(ctx, docs)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) JSONMSet(ctx context.Context, params ...interface{}) *redis.StatusCmd {
	args := r.Called(ctx, params)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) JSONMGet(ctx context.Context, path string, keys ...string) *redis.JSONSliceCmd {
	args := r.Called(ctx, path, keys)
	return args.Get(0).(*redis.JSONSliceCmd)
}

func (r *RedisClient) JSONNumIncrBy(ctx context.Context, key, path string, value float64) *redis.JSONCmd {
	args := r.Called(ctx, key, path, value)
	return args.Get(0).(*redis.JSONCmd)
}

func (r *RedisClient) JSONObjKeys(ctx context.Context, key, path string) *redis.SliceCmd {
	args := r.Called(ctx, key, path)
	return args.Get(0).(*redis.SliceCmd)
}

func (r *RedisClient) JSONObjLen(ctx context.Context, key, path string) *redis.IntPointerSliceCmd {
	args := r.Called(ctx, key, path)
	return args.Get(0).(*redis.IntPointerSliceCmd)
}

func (r *RedisClient) JSONSet(ctx context.Context, key, path string, value interface{}) *redis.StatusCmd {
	args := r.Called(ctx, key, path, value)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) JSONSetMode(ctx context.Context, key, path string, value interface{}, mode string) *redis.StatusCmd {
	args := r.Called(ctx, key, path, value, mode)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) JSONStrAppend(ctx context.Context, key, path, value string) *redis.IntPointerSliceCmd {
	args := r.Called(ctx, key, path, value)
	return args.Get(0).(*redis.IntPointerSliceCmd)
}

func (r *RedisClient) JSONStrLen(ctx context.Context, key, path string) *redis.IntPointerSliceCmd {
	args := r.Called(ctx, key, path)
	return args.Get(0).(*redis.IntPointerSliceCmd)
}

func (r *RedisClient) JSONToggle(ctx context.Context, key, path string) *redis.IntPointerSliceCmd {
	args := r.Called(ctx, key, path)
	return args.Get(0).(*redis.IntPointerSliceCmd)
}

func (r *RedisClient) JSONType(ctx context.Context, key, path string) *redis.JSONSliceCmd {
	args := r.Called(ctx, key, path)
	return args.Get(0).(*redis.JSONSliceCmd)
}

func (r *RedisClient) ACLLog(ctx context.Context, count int64) *redis.ACLLogCmd {
	args := r.Called(ctx, count)
	return args.Get(0).(*redis.ACLLogCmd)
}

func (r *RedisClient) ACLLogReset(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ZRankWithScore(ctx context.Context, key, member string) *redis.RankWithScoreCmd {
	args := r.Called(ctx, key, member)
	return args.Get(0).(*redis.RankWithScoreCmd)
}

func (r *RedisClient) ZRevRankWithScore(ctx context.Context, key, member string) *redis.RankWithScoreCmd {
	args := r.Called(ctx, key, member)
	return args.Get(0).(*redis.RankWithScoreCmd)
}

func (r *RedisClient) ClientInfo(ctx context.Context) *redis.ClientInfoCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.ClientInfoCmd)
}

func (r *RedisClient) FCallRO(ctx context.Context, function string, keys []string, args ...any) *redis.Cmd {
	call := r.Called(ctx, function, keys, args)
	return call.Get(0).(*redis.Cmd)
}

func (r *RedisClient) ClusterMyShardID(ctx context.Context) *redis.StringCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) ModuleLoadex(ctx context.Context, conf *redis.ModuleLoadexConfig) *redis.StringCmd {
	args := r.Called(ctx, conf)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) Pipeline() redis.Pipeliner {
	args := r.Called()
	return args.Get(0).(redis.Pipeliner)
}

func (r *RedisClient) Pipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	args := r.Called(ctx, fn)
	return args.Get(0).([]redis.Cmder), args.Error(1)
}

func (r *RedisClient) TxPipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	args := r.Called(ctx, fn)
	return args.Get(0).([]redis.Cmder), args.Error(1)
}

func (r *RedisClient) TxPipeline() redis.Pipeliner {
	args := r.Called()
	return args.Get(0).(redis.Pipeliner)
}

func (r *RedisClient) Command(ctx context.Context) *redis.CommandsInfoCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.CommandsInfoCmd)
}

func (r *RedisClient) CommandList(ctx context.Context, filter *redis.FilterBy) *redis.StringSliceCmd {
	args := r.Called(ctx, filter)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) CommandGetKeys(ctx context.Context, commands ...any) *redis.StringSliceCmd {
	args := r.Called(ctx, commands)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) CommandGetKeysAndFlags(ctx context.Context, commands ...any) *redis.KeyFlagsCmd {
	args := r.Called(ctx, commands)
	return args.Get(0).(*redis.KeyFlagsCmd)
}

func (r *RedisClient) ClientGetName(ctx context.Context) *redis.StringCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) Echo(ctx context.Context, message any) *redis.StringCmd {
	args := r.Called(ctx, message)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) Quit(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := r.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) Unlink(ctx context.Context, keys ...string) *redis.IntCmd {
	args := r.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) Dump(ctx context.Context, key string) *redis.StringCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	args := r.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := r.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) ExpireAt(ctx context.Context, key string, tm time.Time) *redis.BoolCmd {
	args := r.Called(ctx, key, tm)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) ExpireTime(ctx context.Context, key string) *redis.DurationCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.DurationCmd)
}

func (r *RedisClient) ExpireNX(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := r.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) ExpireXX(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := r.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) ExpireGT(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := r.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) ExpireLT(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := r.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	args := r.Called(ctx, pattern)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) Migrate(ctx context.Context, host, port, key string, db int, timeout time.Duration) *redis.StatusCmd {
	args := r.Called(ctx, host, port, key, db, timeout)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) Move(ctx context.Context, key string, db int) *redis.BoolCmd {
	args := r.Called(ctx, key, db)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) ObjectRefCount(ctx context.Context, key string) *redis.IntCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ObjectEncoding(ctx context.Context, key string) *redis.StringCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) ObjectIdleTime(ctx context.Context, key string) *redis.DurationCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.DurationCmd)
}

func (r *RedisClient) Persist(ctx context.Context, key string) *redis.BoolCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) PExpire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := r.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) PExpireAt(ctx context.Context, key string, tm time.Time) *redis.BoolCmd {
	args := r.Called(ctx, key, tm)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) PExpireTime(ctx context.Context, key string) *redis.DurationCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.DurationCmd)
}

func (r *RedisClient) PTTL(ctx context.Context, key string) *redis.DurationCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.DurationCmd)
}

func (r *RedisClient) RandomKey(ctx context.Context) *redis.StringCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) Rename(ctx context.Context, key, newkey string) *redis.StatusCmd {
	args := r.Called(ctx, key, newkey)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) RenameNX(ctx context.Context, key, newkey string) *redis.BoolCmd {
	args := r.Called(ctx, key, newkey)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) Restore(ctx context.Context, key string, ttl time.Duration, value string) *redis.StatusCmd {
	args := r.Called(ctx, key, ttl, value)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) RestoreReplace(ctx context.Context, key string, ttl time.Duration, value string) *redis.StatusCmd {
	args := r.Called(ctx, key, ttl, value)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) Sort(ctx context.Context, key string, sort *redis.Sort) *redis.StringSliceCmd {
	args := r.Called(ctx, key, sort)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) SortRO(ctx context.Context, key string, sort *redis.Sort) *redis.StringSliceCmd {
	args := r.Called(ctx, key, sort)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) SortStore(ctx context.Context, key, store string, sort *redis.Sort) *redis.IntCmd {
	args := r.Called(ctx, key, store, sort)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) SortInterfaces(ctx context.Context, key string, sort *redis.Sort) *redis.SliceCmd {
	args := r.Called(ctx, key, sort)
	return args.Get(0).(*redis.SliceCmd)
}

func (r *RedisClient) Touch(ctx context.Context, keys ...string) *redis.IntCmd {
	args := r.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) TTL(ctx context.Context, key string) *redis.DurationCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.DurationCmd)
}

func (r *RedisClient) Type(ctx context.Context, key string) *redis.StatusCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) Append(ctx context.Context, key, value string) *redis.IntCmd {
	args := r.Called(ctx, key, value)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) Decr(ctx context.Context, key string) *redis.IntCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) DecrBy(ctx context.Context, key string, decrement int64) *redis.IntCmd {
	args := r.Called(ctx, key, decrement)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) GetRange(ctx context.Context, key string, start, end int64) *redis.StringCmd {
	args := r.Called(ctx, key, start, end)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) GetSet(ctx context.Context, key string, value any) *redis.StringCmd {
	args := r.Called(ctx, key, value)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) GetEx(ctx context.Context, key string, expiration time.Duration) *redis.StringCmd {
	args := r.Called(ctx, key, expiration)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) GetDel(ctx context.Context, key string) *redis.StringCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) IncrBy(ctx context.Context, key string, value int64) *redis.IntCmd {
	args := r.Called(ctx, key, value)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) IncrByFloat(ctx context.Context, key string, value float64) *redis.FloatCmd {
	args := r.Called(ctx, key, value)
	return args.Get(0).(*redis.FloatCmd)
}

func (r *RedisClient) MGet(ctx context.Context, keys ...string) *redis.SliceCmd {
	args := r.Called(ctx, keys)
	return args.Get(0).(*redis.SliceCmd)
}

func (r *RedisClient) MSet(ctx context.Context, values ...any) *redis.StatusCmd {
	args := r.Called(ctx, values)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) MSetNX(ctx context.Context, values ...any) *redis.BoolCmd {
	args := r.Called(ctx, values)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	args := r.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) SetArgs(ctx context.Context, key string, value any, a redis.SetArgs) *redis.StatusCmd {
	args := r.Called(ctx, key, value, a)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) SetEx(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	args := r.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) SetNX(ctx context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd {
	args := r.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) SetXX(ctx context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd {
	args := r.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) SetRange(ctx context.Context, key string, offset int64, value string) *redis.IntCmd {
	args := r.Called(ctx, key, offset, value)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) StrLen(ctx context.Context, key string) *redis.IntCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) Copy(ctx context.Context, sourceKey string, destKey string, db int, replace bool) *redis.IntCmd {
	args := r.Called(ctx, sourceKey, destKey, db, replace)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) GetBit(ctx context.Context, key string, offset int64) *redis.IntCmd {
	args := r.Called(ctx, key, offset)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) SetBit(ctx context.Context, key string, offset int64, value int) *redis.IntCmd {
	args := r.Called(ctx, key, offset, value)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) BitCount(ctx context.Context, key string, bitCount *redis.BitCount) *redis.IntCmd {
	args := r.Called(ctx, key, bitCount)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) BitOpAnd(ctx context.Context, destKey string, keys ...string) *redis.IntCmd {
	args := r.Called(ctx, destKey, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) BitOpOr(ctx context.Context, destKey string, keys ...string) *redis.IntCmd {
	args := r.Called(ctx, destKey, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) BitOpXor(ctx context.Context, destKey string, keys ...string) *redis.IntCmd {
	args := r.Called(ctx, destKey, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) BitOpNot(ctx context.Context, destKey string, key string) *redis.IntCmd {
	args := r.Called(ctx, destKey, key)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) BitPos(ctx context.Context, key string, bit int64, pos ...int64) *redis.IntCmd {
	args := r.Called(ctx, key, bit, pos)
	return args.Get(0).(*redis.IntCmd)

}

func (r *RedisClient) BitPosSpan(ctx context.Context, key string, bit int8, start, end int64, span string) *redis.IntCmd {
	args := r.Called(ctx, key, bit, start, end, span)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) BitField(ctx context.Context, key string, args ...any) *redis.IntSliceCmd {
	call := r.Called(ctx, key, args)
	return call.Get(0).(*redis.IntSliceCmd)
}

func (r *RedisClient) BitFieldRO(ctx context.Context, key string, values ...any) *redis.IntSliceCmd {
	args := r.Called(ctx, key, values)
	return args.Get(0).(*redis.IntSliceCmd)
}

func (r *RedisClient) Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	args := r.Called(ctx, cursor, match, count)
	return args.Get(0).(*redis.ScanCmd)
}

func (r *RedisClient) ScanType(ctx context.Context, cursor uint64, match string, count int64, keyType string) *redis.ScanCmd {
	args := r.Called(ctx, cursor, match, count, keyType)
	return args.Get(0).(*redis.ScanCmd)
}

func (r *RedisClient) SScan(ctx context.Context, key string, cursor uint64, match string, count int64) *redis.ScanCmd {
	args := r.Called(ctx, key, cursor, match, count)
	return args.Get(0).(*redis.ScanCmd)
}

func (r *RedisClient) HScan(ctx context.Context, key string, cursor uint64, match string, count int64) *redis.ScanCmd {
	args := r.Called(ctx, key, cursor, match, count)
	return args.Get(0).(*redis.ScanCmd)
}

func (r *RedisClient) ZScan(ctx context.Context, key string, cursor uint64, match string, count int64) *redis.ScanCmd {
	args := r.Called(ctx, key, cursor, match, count)
	return args.Get(0).(*redis.ScanCmd)
}

func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	args := r.Called(ctx, key, fields)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) HExists(ctx context.Context, key, field string) *redis.BoolCmd {
	args := r.Called(ctx, key, field)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	args := r.Called(ctx, key, field)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.MapStringStringCmd)
}

func (r *RedisClient) HIncrBy(ctx context.Context, key, field string, incr int64) *redis.IntCmd {
	args := r.Called(ctx, key, field, incr)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) HIncrByFloat(ctx context.Context, key, field string, incr float64) *redis.FloatCmd {
	args := r.Called(ctx, key, field, incr)
	return args.Get(0).(*redis.FloatCmd)
}

func (r *RedisClient) HKeys(ctx context.Context, key string) *redis.StringSliceCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) HLen(ctx context.Context, key string) *redis.IntCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) HMGet(ctx context.Context, key string, fields ...string) *redis.SliceCmd {
	args := r.Called(ctx, key, fields)
	return args.Get(0).(*redis.SliceCmd)
}

func (r *RedisClient) HSet(ctx context.Context, key string, values ...any) *redis.IntCmd {
	args := r.Called(ctx, key, values)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) HMSet(ctx context.Context, key string, values ...any) *redis.BoolCmd {
	args := r.Called(ctx, key, values)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) HSetNX(ctx context.Context, key, field string, value any) *redis.BoolCmd {
	args := r.Called(ctx, key, field, value)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) HVals(ctx context.Context, key string) *redis.StringSliceCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) HRandField(ctx context.Context, key string, count int) *redis.StringSliceCmd {
	args := r.Called(ctx, key, count)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) HRandFieldWithValues(ctx context.Context, key string, count int) *redis.KeyValueSliceCmd {
	args := r.Called(ctx, key, count)
	return args.Get(0).(*redis.KeyValueSliceCmd)
}

func (r *RedisClient) BLPop(ctx context.Context, timeout time.Duration, keys ...string) *redis.StringSliceCmd {
	args := r.Called(ctx, timeout, keys)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) BLMPop(ctx context.Context, timeout time.Duration, direction string, count int64, keys ...string) *redis.KeyValuesCmd {
	args := r.Called(ctx, timeout, direction, count, keys)
	return args.Get(0).(*redis.KeyValuesCmd)
}

func (r *RedisClient) BRPop(ctx context.Context, timeout time.Duration, keys ...string) *redis.StringSliceCmd {
	args := r.Called(ctx, timeout, keys)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) BRPopLPush(ctx context.Context, source, destination string, timeout time.Duration) *redis.StringCmd {
	args := r.Called(ctx, source, destination, timeout)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) LCS(ctx context.Context, q *redis.LCSQuery) *redis.LCSCmd {
	args := r.Called(ctx, q)
	return args.Get(0).(*redis.LCSCmd)
}

func (r *RedisClient) LIndex(ctx context.Context, key string, index int64) *redis.StringCmd {
	args := r.Called(ctx, key, index)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) LInsert(ctx context.Context, key, op string, pivot, value any) *redis.IntCmd {
	args := r.Called(ctx, key, op, pivot, value)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) LInsertBefore(ctx context.Context, key string, pivot, value any) *redis.IntCmd {
	args := r.Called(ctx, key, pivot, value)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) LInsertAfter(ctx context.Context, key string, pivot, value any) *redis.IntCmd {
	args := r.Called(ctx, key, pivot, value)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) LLen(ctx context.Context, key string) *redis.IntCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) LMPop(ctx context.Context, direction string, count int64, keys ...string) *redis.KeyValuesCmd {
	args := r.Called(ctx, direction, count, keys)
	return args.Get(0).(*redis.KeyValuesCmd)
}

func (r *RedisClient) LPop(ctx context.Context, key string) *redis.StringCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) LPopCount(ctx context.Context, key string, count int) *redis.StringSliceCmd {
	args := r.Called(ctx, key, count)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) LPos(ctx context.Context, key string, value string, args redis.LPosArgs) *redis.IntCmd {
	call := r.Called(ctx, key, value, args)
	return call.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) LPosCount(ctx context.Context, key string, value string, count int64, args redis.LPosArgs) *redis.IntSliceCmd {
	call := r.Called(ctx, key, value, count, args)
	return call.Get(0).(*redis.IntSliceCmd)
}

func (r *RedisClient) LPush(ctx context.Context, key string, values ...any) *redis.IntCmd {
	args := r.Called(ctx, key, values)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) LPushX(ctx context.Context, key string, values ...any) *redis.IntCmd {
	args := r.Called(ctx, key, values)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) LRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	args := r.Called(ctx, key, start, stop)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) LRem(ctx context.Context, key string, count int64, value any) *redis.IntCmd {
	args := r.Called(ctx, key, count, value)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) LSet(ctx context.Context, key string, index int64, value any) *redis.StatusCmd {
	args := r.Called(ctx, key, index, value)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) LTrim(ctx context.Context, key string, start, stop int64) *redis.StatusCmd {
	args := r.Called(ctx, key, start, stop)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) RPop(ctx context.Context, key string) *redis.StringCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) RPopCount(ctx context.Context, key string, count int) *redis.StringSliceCmd {
	args := r.Called(ctx, key, count)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) RPopLPush(ctx context.Context, source, destination string) *redis.StringCmd {
	args := r.Called(ctx, source, destination)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) RPush(ctx context.Context, key string, values ...any) *redis.IntCmd {
	args := r.Called(ctx, key, values)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) RPushX(ctx context.Context, key string, values ...any) *redis.IntCmd {
	args := r.Called(ctx, key, values)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) LMove(ctx context.Context, source, destination, srcpos, destpos string) *redis.StringCmd {
	args := r.Called(ctx, source, destination, srcpos, destpos)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) BLMove(ctx context.Context, source, destination, srcpos, destpos string, timeout time.Duration) *redis.StringCmd {
	args := r.Called(ctx, source, destination, srcpos, destpos, timeout)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) SAdd(ctx context.Context, key string, members ...any) *redis.IntCmd {
	args := r.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) SCard(ctx context.Context, key string) *redis.IntCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) SDiff(ctx context.Context, keys ...string) *redis.StringSliceCmd {
	args := r.Called(ctx, keys)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) SDiffStore(ctx context.Context, destination string, keys ...string) *redis.IntCmd {
	args := r.Called(ctx, destination, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) SInter(ctx context.Context, keys ...string) *redis.StringSliceCmd {
	args := r.Called(ctx, keys)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) SInterCard(ctx context.Context, limit int64, keys ...string) *redis.IntCmd {
	args := r.Called(ctx, limit, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) SInterStore(ctx context.Context, destination string, keys ...string) *redis.IntCmd {
	args := r.Called(ctx, destination, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) SIsMember(ctx context.Context, key string, member any) *redis.BoolCmd {
	args := r.Called(ctx, key, member)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) SMIsMember(ctx context.Context, key string, members ...any) *redis.BoolSliceCmd {
	args := r.Called(ctx, key, members)
	return args.Get(0).(*redis.BoolSliceCmd)
}

func (r *RedisClient) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) SMembersMap(ctx context.Context, key string) *redis.StringStructMapCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.StringStructMapCmd)
}

func (r *RedisClient) SMove(ctx context.Context, source, destination string, member any) *redis.BoolCmd {
	args := r.Called(ctx, source, destination, member)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) SPop(ctx context.Context, key string) *redis.StringCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) SPopN(ctx context.Context, key string, count int64) *redis.StringSliceCmd {
	args := r.Called(ctx, key, count)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) SRandMember(ctx context.Context, key string) *redis.StringCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) SRandMemberN(ctx context.Context, key string, count int64) *redis.StringSliceCmd {
	args := r.Called(ctx, key, count)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) SRem(ctx context.Context, key string, members ...any) *redis.IntCmd {
	args := r.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) SUnion(ctx context.Context, keys ...string) *redis.StringSliceCmd {
	args := r.Called(ctx, keys)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) SUnionStore(ctx context.Context, destination string, keys ...string) *redis.IntCmd {
	args := r.Called(ctx, destination, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) XAdd(ctx context.Context, a *redis.XAddArgs) *redis.StringCmd {
	args := r.Called(ctx, a)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) XDel(ctx context.Context, stream string, ids ...string) *redis.IntCmd {
	args := r.Called(ctx, stream, ids)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) XLen(ctx context.Context, stream string) *redis.IntCmd {
	args := r.Called(ctx, stream)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) XRange(ctx context.Context, stream, start, stop string) *redis.XMessageSliceCmd {
	args := r.Called(ctx, stream, start, stop)
	return args.Get(0).(*redis.XMessageSliceCmd)
}

func (r *RedisClient) XRangeN(ctx context.Context, stream, start, stop string, count int64) *redis.XMessageSliceCmd {
	args := r.Called(ctx, stream, start, stop, count)
	return args.Get(0).(*redis.XMessageSliceCmd)
}

func (r *RedisClient) XRevRange(ctx context.Context, stream string, start, stop string) *redis.XMessageSliceCmd {
	args := r.Called(ctx, stream, start, stop)
	return args.Get(0).(*redis.XMessageSliceCmd)
}

func (r *RedisClient) XRevRangeN(ctx context.Context, stream string, start, stop string, count int64) *redis.XMessageSliceCmd {
	args := r.Called(ctx, stream, start, stop, count)
	return args.Get(0).(*redis.XMessageSliceCmd)
}

func (r *RedisClient) XRead(ctx context.Context, a *redis.XReadArgs) *redis.XStreamSliceCmd {
	args := r.Called(ctx, a)
	return args.Get(0).(*redis.XStreamSliceCmd)
}

func (r *RedisClient) XReadStreams(ctx context.Context, streams ...string) *redis.XStreamSliceCmd {
	args := r.Called(ctx, streams)
	return args.Get(0).(*redis.XStreamSliceCmd)
}

func (r *RedisClient) XGroupCreate(ctx context.Context, stream, group, start string) *redis.StatusCmd {
	args := r.Called(ctx, stream, group, start)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) XGroupCreateMkStream(ctx context.Context, stream, group, start string) *redis.StatusCmd {
	args := r.Called(ctx, stream, group, start)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) XGroupSetID(ctx context.Context, stream, group, start string) *redis.StatusCmd {
	args := r.Called(ctx, stream, group, start)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) XGroupDestroy(ctx context.Context, stream, group string) *redis.IntCmd {
	args := r.Called(ctx, stream, group)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) XGroupCreateConsumer(ctx context.Context, stream, group, consumer string) *redis.IntCmd {
	args := r.Called(ctx, stream, group, consumer)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) XGroupDelConsumer(ctx context.Context, stream, group, consumer string) *redis.IntCmd {
	args := r.Called(ctx, stream, group, consumer)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) XReadGroup(ctx context.Context, a *redis.XReadGroupArgs) *redis.XStreamSliceCmd {
	args := r.Called(ctx, a)
	return args.Get(0).(*redis.XStreamSliceCmd)
}

func (r *RedisClient) XAck(ctx context.Context, stream, group string, ids ...string) *redis.IntCmd {
	args := r.Called(ctx, stream, group, ids)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) XPending(ctx context.Context, stream, group string) *redis.XPendingCmd {
	args := r.Called(ctx, stream, group)
	return args.Get(0).(*redis.XPendingCmd)
}

func (r *RedisClient) XPendingExt(ctx context.Context, a *redis.XPendingExtArgs) *redis.XPendingExtCmd {
	args := r.Called(ctx, a)
	return args.Get(0).(*redis.XPendingExtCmd)
}

func (r *RedisClient) XClaim(ctx context.Context, a *redis.XClaimArgs) *redis.XMessageSliceCmd {
	args := r.Called(ctx, a)
	return args.Get(0).(*redis.XMessageSliceCmd)
}

func (r *RedisClient) XClaimJustID(ctx context.Context, a *redis.XClaimArgs) *redis.StringSliceCmd {
	args := r.Called(ctx, a)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) XAutoClaim(ctx context.Context, a *redis.XAutoClaimArgs) *redis.XAutoClaimCmd {
	args := r.Called(ctx, a)
	return args.Get(0).(*redis.XAutoClaimCmd)
}

func (r *RedisClient) XAutoClaimJustID(ctx context.Context, a *redis.XAutoClaimArgs) *redis.XAutoClaimJustIDCmd {
	args := r.Called(ctx, a)
	return args.Get(0).(*redis.XAutoClaimJustIDCmd)
}

func (r *RedisClient) XTrimMaxLen(ctx context.Context, key string, maxLen int64) *redis.IntCmd {
	args := r.Called(ctx, key, maxLen)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) XTrimMaxLenApprox(ctx context.Context, key string, maxLen, limit int64) *redis.IntCmd {
	args := r.Called(ctx, key, maxLen, limit)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) XTrimMinID(ctx context.Context, key string, minID string) *redis.IntCmd {
	args := r.Called(ctx, key, minID)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) XTrimMinIDApprox(ctx context.Context, key string, minID string, limit int64) *redis.IntCmd {
	args := r.Called(ctx, key, minID, limit)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) XInfoGroups(ctx context.Context, key string) *redis.XInfoGroupsCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.XInfoGroupsCmd)
}

func (r *RedisClient) XInfoStream(ctx context.Context, key string) *redis.XInfoStreamCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.XInfoStreamCmd)
}

func (r *RedisClient) XInfoStreamFull(ctx context.Context, key string, count int) *redis.XInfoStreamFullCmd {
	args := r.Called(ctx, key, count)
	return args.Get(0).(*redis.XInfoStreamFullCmd)
}

func (r *RedisClient) XInfoConsumers(ctx context.Context, key string, group string) *redis.XInfoConsumersCmd {
	args := r.Called(ctx, key, group)
	return args.Get(0).(*redis.XInfoConsumersCmd)
}

func (r *RedisClient) BZPopMax(ctx context.Context, timeout time.Duration, keys ...string) *redis.ZWithKeyCmd {
	args := r.Called(ctx, timeout, keys)
	return args.Get(0).(*redis.ZWithKeyCmd)
}

func (r *RedisClient) BZPopMin(ctx context.Context, timeout time.Duration, keys ...string) *redis.ZWithKeyCmd {
	args := r.Called(ctx, timeout, keys)
	return args.Get(0).(*redis.ZWithKeyCmd)
}

func (r *RedisClient) BZMPop(ctx context.Context, timeout time.Duration, order string, count int64, keys ...string) *redis.ZSliceWithKeyCmd {
	args := r.Called(ctx, timeout, order, count, keys)
	return args.Get(0).(*redis.ZSliceWithKeyCmd)
}

func (r *RedisClient) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	args := r.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ZAddLT(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	args := r.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ZAddGT(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	args := r.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ZAddNX(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	args := r.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ZAddXX(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	args := r.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ZAddArgs(ctx context.Context, key string, args redis.ZAddArgs) *redis.IntCmd {
	call := r.Called(ctx, key, args)
	return call.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ZAddArgsIncr(ctx context.Context, key string, args redis.ZAddArgs) *redis.FloatCmd {
	call := r.Called(ctx, key, args)
	return call.Get(0).(*redis.FloatCmd)
}

func (r *RedisClient) ZCard(ctx context.Context, key string) *redis.IntCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ZCount(ctx context.Context, key, _min, _max string) *redis.IntCmd {
	args := r.Called(ctx, key, _min, _max)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ZLexCount(ctx context.Context, key, _min, _max string) *redis.IntCmd {
	args := r.Called(ctx, key, _min, _max)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ZIncrBy(ctx context.Context, key string, increment float64, member string) *redis.FloatCmd {
	args := r.Called(ctx, key, increment, member)
	return args.Get(0).(*redis.FloatCmd)
}

func (r *RedisClient) ZInter(ctx context.Context, store *redis.ZStore) *redis.StringSliceCmd {
	args := r.Called(ctx, store)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) ZInterWithScores(ctx context.Context, store *redis.ZStore) *redis.ZSliceCmd {
	args := r.Called(ctx, store)
	return args.Get(0).(*redis.ZSliceCmd)
}

func (r *RedisClient) ZInterCard(ctx context.Context, limit int64, keys ...string) *redis.IntCmd {
	args := r.Called(ctx, limit, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ZInterStore(ctx context.Context, destination string, store *redis.ZStore) *redis.IntCmd {
	args := r.Called(ctx, destination, store)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ZMPop(ctx context.Context, order string, count int64, keys ...string) *redis.ZSliceWithKeyCmd {
	args := r.Called(ctx, order, count, keys)
	return args.Get(0).(*redis.ZSliceWithKeyCmd)
}

func (r *RedisClient) ZMScore(ctx context.Context, key string, members ...string) *redis.FloatSliceCmd {
	args := r.Called(ctx, key, members)
	return args.Get(0).(*redis.FloatSliceCmd)
}

func (r *RedisClient) ZPopMax(ctx context.Context, key string, count ...int64) *redis.ZSliceCmd {
	args := r.Called(ctx, key, count)
	return args.Get(0).(*redis.ZSliceCmd)
}

func (r *RedisClient) ZPopMin(ctx context.Context, key string, count ...int64) *redis.ZSliceCmd {
	args := r.Called(ctx, key, count)
	return args.Get(0).(*redis.ZSliceCmd)
}

func (r *RedisClient) ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	args := r.Called(ctx, key, start, stop)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) ZRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd {
	args := r.Called(ctx, key, start, stop)
	return args.Get(0).(*redis.ZSliceCmd)
}

func (r *RedisClient) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	args := r.Called(ctx, key, opt)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) ZRangeByLex(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	args := r.Called(ctx, key, opt)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) ZRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.ZSliceCmd {
	args := r.Called(ctx, key, opt)
	return args.Get(0).(*redis.ZSliceCmd)
}

func (r *RedisClient) ZRangeArgs(ctx context.Context, z redis.ZRangeArgs) *redis.StringSliceCmd {
	args := r.Called(ctx, z)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) ZRangeArgsWithScores(ctx context.Context, z redis.ZRangeArgs) *redis.ZSliceCmd {
	args := r.Called(ctx, z)
	return args.Get(0).(*redis.ZSliceCmd)
}

func (r *RedisClient) ZRangeStore(ctx context.Context, dst string, z redis.ZRangeArgs) *redis.IntCmd {
	args := r.Called(ctx, dst, z)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ZRank(ctx context.Context, key, member string) *redis.IntCmd {
	args := r.Called(ctx, key, member)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ZRem(ctx context.Context, key string, members ...any) *redis.IntCmd {
	args := r.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) *redis.IntCmd {
	args := r.Called(ctx, key, start, stop)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ZRemRangeByScore(ctx context.Context, key, _min, _max string) *redis.IntCmd {
	args := r.Called(ctx, key, _min, _max)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ZRemRangeByLex(ctx context.Context, key, _min, _max string) *redis.IntCmd {
	args := r.Called(ctx, key, _min, _max)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ZRevRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	args := r.Called(ctx, key, start, stop)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd {
	args := r.Called(ctx, key, start, stop)
	return args.Get(0).(*redis.ZSliceCmd)
}

func (r *RedisClient) ZRevRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	args := r.Called(ctx, key, opt)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) ZRevRangeByLex(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	args := r.Called(ctx, key, opt)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) ZRevRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.ZSliceCmd {
	args := r.Called(ctx, key, opt)
	return args.Get(0).(*redis.ZSliceCmd)
}

func (r *RedisClient) ZRevRank(ctx context.Context, key, member string) *redis.IntCmd {
	args := r.Called(ctx, key, member)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ZScore(ctx context.Context, key, member string) *redis.FloatCmd {
	args := r.Called(ctx, key, member)
	return args.Get(0).(*redis.FloatCmd)
}

func (r *RedisClient) ZUnionStore(ctx context.Context, dest string, store *redis.ZStore) *redis.IntCmd {
	args := r.Called(ctx, dest, store)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ZRandMember(ctx context.Context, key string, count int) *redis.StringSliceCmd {
	args := r.Called(ctx, key, count)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) ZRandMemberWithScores(ctx context.Context, key string, count int) *redis.ZSliceCmd {
	args := r.Called(ctx, key, count)
	return args.Get(0).(*redis.ZSliceCmd)
}

func (r *RedisClient) ZUnion(ctx context.Context, store redis.ZStore) *redis.StringSliceCmd {
	args := r.Called(ctx, store)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) ZUnionWithScores(ctx context.Context, store redis.ZStore) *redis.ZSliceCmd {
	args := r.Called(ctx, store)
	return args.Get(0).(*redis.ZSliceCmd)
}

func (r *RedisClient) ZDiff(ctx context.Context, keys ...string) *redis.StringSliceCmd {
	args := r.Called(ctx, keys)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) ZDiffWithScores(ctx context.Context, keys ...string) *redis.ZSliceCmd {
	args := r.Called(ctx, keys)
	return args.Get(0).(*redis.ZSliceCmd)
}

func (r *RedisClient) ZDiffStore(ctx context.Context, destination string, keys ...string) *redis.IntCmd {
	args := r.Called(ctx, destination, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) PFAdd(ctx context.Context, key string, els ...any) *redis.IntCmd {
	args := r.Called(ctx, key, els)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) PFCount(ctx context.Context, keys ...string) *redis.IntCmd {
	args := r.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) PFMerge(ctx context.Context, dest string, keys ...string) *redis.StatusCmd {
	args := r.Called(ctx, dest, keys)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) BgRewriteAOF(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) BgSave(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ClientKill(ctx context.Context, ipPort string) *redis.StatusCmd {
	args := r.Called(ctx, ipPort)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ClientKillByFilter(ctx context.Context, keys ...string) *redis.IntCmd {
	args := r.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ClientList(ctx context.Context) *redis.StringCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) ClientPause(ctx context.Context, dur time.Duration) *redis.BoolCmd {
	args := r.Called(ctx, dur)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) ClientUnpause(ctx context.Context) *redis.BoolCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.BoolCmd)
}

func (r *RedisClient) ClientID(ctx context.Context) *redis.IntCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ClientUnblock(ctx context.Context, id int64) *redis.IntCmd {
	args := r.Called(ctx, id)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ClientUnblockWithError(ctx context.Context, id int64) *redis.IntCmd {
	args := r.Called(ctx, id)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ConfigGet(ctx context.Context, parameter string) *redis.MapStringStringCmd {
	args := r.Called(ctx, parameter)
	return args.Get(0).(*redis.MapStringStringCmd)
}

func (r *RedisClient) ConfigResetStat(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ConfigSet(ctx context.Context, parameter, value string) *redis.StatusCmd {
	args := r.Called(ctx, parameter, value)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ConfigRewrite(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) DBSize(ctx context.Context) *redis.IntCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) FlushAll(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) FlushAllAsync(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) FlushDB(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) FlushDBAsync(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) Info(ctx context.Context, section ...string) *redis.StringCmd {
	args := r.Called(ctx, section)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) LastSave(ctx context.Context) *redis.IntCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) Save(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) Shutdown(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ShutdownSave(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ShutdownNoSave(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) SlaveOf(ctx context.Context, host, port string) *redis.StatusCmd {
	args := r.Called(ctx, host, port)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) SlowLogGet(ctx context.Context, num int64) *redis.SlowLogCmd {
	args := r.Called(ctx, num)
	return args.Get(0).(*redis.SlowLogCmd)
}

func (r *RedisClient) Time(ctx context.Context) *redis.TimeCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.TimeCmd)
}

func (r *RedisClient) DebugObject(ctx context.Context, key string) *redis.StringCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) ReadOnly(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ReadWrite(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) MemoryUsage(ctx context.Context, key string, samples ...int) *redis.IntCmd {
	args := r.Called(ctx, key, samples)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) Eval(ctx context.Context, script string, keys []string, args ...any) *redis.Cmd {
	call := r.Called(ctx, script, keys, args)
	return call.Get(0).(*redis.Cmd)
}

func (r *RedisClient) EvalSha(ctx context.Context, sha1 string, keys []string, args ...any) *redis.Cmd {
	call := r.Called(ctx, sha1, keys, args)
	return call.Get(0).(*redis.Cmd)
}

func (r *RedisClient) EvalRO(ctx context.Context, script string, keys []string, args ...any) *redis.Cmd {
	call := r.Called(ctx, script, keys, args)
	return call.Get(0).(*redis.Cmd)
}

func (r *RedisClient) EvalShaRO(ctx context.Context, sha1 string, keys []string, args ...any) *redis.Cmd {
	call := r.Called(ctx, sha1, keys, args)
	return call.Get(0).(*redis.Cmd)
}

func (r *RedisClient) ScriptExists(ctx context.Context, hashes ...string) *redis.BoolSliceCmd {
	args := r.Called(ctx, hashes)
	return args.Get(0).(*redis.BoolSliceCmd)
}

func (r *RedisClient) ScriptFlush(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ScriptKill(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ScriptLoad(ctx context.Context, script string) *redis.StringCmd {
	args := r.Called(ctx, script)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) FunctionLoad(ctx context.Context, code string) *redis.StringCmd {
	args := r.Called(ctx, code)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) FunctionLoadReplace(ctx context.Context, code string) *redis.StringCmd {
	args := r.Called(ctx, code)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) FunctionDelete(ctx context.Context, libName string) *redis.StringCmd {
	args := r.Called(ctx, libName)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) FunctionFlush(ctx context.Context) *redis.StringCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) FunctionKill(ctx context.Context) *redis.StringCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) FunctionFlushAsync(ctx context.Context) *redis.StringCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) FunctionList(ctx context.Context, q redis.FunctionListQuery) *redis.FunctionListCmd {
	args := r.Called(ctx, q)
	return args.Get(0).(*redis.FunctionListCmd)
}

func (r *RedisClient) FunctionDump(ctx context.Context) *redis.StringCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) FunctionRestore(ctx context.Context, libDump string) *redis.StringCmd {
	args := r.Called(ctx, libDump)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) FunctionStats(ctx context.Context) *redis.FunctionStatsCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.FunctionStatsCmd)
}

func (r *RedisClient) FCall(ctx context.Context, function string, keys []string, args ...any) *redis.Cmd {
	call := r.Called(ctx, function, keys, args)
	return call.Get(0).(*redis.Cmd)
}

func (r *RedisClient) FCallRo(ctx context.Context, function string, keys []string, args ...any) *redis.Cmd {
	call := r.Called(ctx, function, keys, args)
	return call.Get(0).(*redis.Cmd)
}

func (r *RedisClient) Publish(ctx context.Context, channel string, message any) *redis.IntCmd {
	args := r.Called(ctx, channel, message)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) SPublish(ctx context.Context, channel string, message any) *redis.IntCmd {
	args := r.Called(ctx, channel, message)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) PubSubChannels(ctx context.Context, pattern string) *redis.StringSliceCmd {
	args := r.Called(ctx, pattern)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) PubSubNumSub(ctx context.Context, channels ...string) *redis.MapStringIntCmd {
	args := r.Called(ctx, channels)
	return args.Get(0).(*redis.MapStringIntCmd)
}

func (r *RedisClient) PubSubNumPat(ctx context.Context) *redis.IntCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) PubSubShardChannels(ctx context.Context, pattern string) *redis.StringSliceCmd {
	args := r.Called(ctx, pattern)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) PubSubShardNumSub(ctx context.Context, channels ...string) *redis.MapStringIntCmd {
	args := r.Called(ctx, channels)
	return args.Get(0).(*redis.MapStringIntCmd)
}

func (r *RedisClient) ClusterSlots(ctx context.Context) *redis.ClusterSlotsCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.ClusterSlotsCmd)
}

func (r *RedisClient) ClusterShards(ctx context.Context) *redis.ClusterShardsCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.ClusterShardsCmd)
}

func (r *RedisClient) ClusterLinks(ctx context.Context) *redis.ClusterLinksCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.ClusterLinksCmd)
}

func (r *RedisClient) ClusterNodes(ctx context.Context) *redis.StringCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) ClusterMeet(ctx context.Context, host, port string) *redis.StatusCmd {
	args := r.Called(ctx, host, port)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ClusterForget(ctx context.Context, nodeID string) *redis.StatusCmd {
	args := r.Called(ctx, nodeID)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ClusterReplicate(ctx context.Context, nodeID string) *redis.StatusCmd {
	args := r.Called(ctx, nodeID)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ClusterResetSoft(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ClusterResetHard(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ClusterInfo(ctx context.Context) *redis.StringCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) ClusterKeySlot(ctx context.Context, key string) *redis.IntCmd {
	args := r.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ClusterGetKeysInSlot(ctx context.Context, slot int, count int) *redis.StringSliceCmd {
	args := r.Called(ctx, slot, count)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) ClusterCountFailureReports(ctx context.Context, nodeID string) *redis.IntCmd {
	args := r.Called(ctx, nodeID)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ClusterCountKeysInSlot(ctx context.Context, slot int) *redis.IntCmd {
	args := r.Called(ctx, slot)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) ClusterDelSlots(ctx context.Context, slots ...int) *redis.StatusCmd {
	args := r.Called(ctx, slots)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ClusterDelSlotsRange(ctx context.Context, _min, _max int) *redis.StatusCmd {
	args := r.Called(ctx, _min, _max)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ClusterSaveConfig(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ClusterSlaves(ctx context.Context, nodeID string) *redis.StringSliceCmd {
	args := r.Called(ctx, nodeID)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) ClusterFailover(ctx context.Context) *redis.StatusCmd {
	args := r.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ClusterAddSlots(ctx context.Context, slots ...int) *redis.StatusCmd {
	args := r.Called(ctx, slots)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) ClusterAddSlotsRange(ctx context.Context, _min, _max int) *redis.StatusCmd {
	args := r.Called(ctx, _min, _max)
	return args.Get(0).(*redis.StatusCmd)
}

func (r *RedisClient) GeoAdd(ctx context.Context, key string, geoLocation ...*redis.GeoLocation) *redis.IntCmd {
	args := r.Called(ctx, key, geoLocation)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) GeoPos(ctx context.Context, key string, members ...string) *redis.GeoPosCmd {
	args := r.Called(ctx, key, members)
	return args.Get(0).(*redis.GeoPosCmd)
}

func (r *RedisClient) GeoRadius(ctx context.Context, key string, longitude, latitude float64, query *redis.GeoRadiusQuery) *redis.GeoLocationCmd {
	args := r.Called(ctx, key, longitude, latitude, query)
	return args.Get(0).(*redis.GeoLocationCmd)
}

func (r *RedisClient) GeoRadiusStore(ctx context.Context, key string, longitude, latitude float64, query *redis.GeoRadiusQuery) *redis.IntCmd {
	args := r.Called(ctx, key, longitude, latitude, query)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) GeoRadiusByMember(ctx context.Context, key, member string, query *redis.GeoRadiusQuery) *redis.GeoLocationCmd {
	args := r.Called(ctx, key, member, query)
	return args.Get(0).(*redis.GeoLocationCmd)
}

func (r *RedisClient) GeoRadiusByMemberStore(ctx context.Context, key, member string, query *redis.GeoRadiusQuery) *redis.IntCmd {
	args := r.Called(ctx, key, member, query)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) GeoSearch(ctx context.Context, key string, q *redis.GeoSearchQuery) *redis.StringSliceCmd {
	args := r.Called(ctx, key, q)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) GeoSearchLocation(ctx context.Context, key string, q *redis.GeoSearchLocationQuery) *redis.GeoSearchLocationCmd {
	args := r.Called(ctx, key, q)
	return args.Get(0).(*redis.GeoSearchLocationCmd)
}

func (r *RedisClient) GeoSearchStore(ctx context.Context, key, store string, q *redis.GeoSearchStoreQuery) *redis.IntCmd {
	args := r.Called(ctx, key, store, q)
	return args.Get(0).(*redis.IntCmd)
}

func (r *RedisClient) GeoDist(ctx context.Context, key string, member1, member2, unit string) *redis.FloatCmd {
	args := r.Called(ctx, key, member1, member2, unit)
	return args.Get(0).(*redis.FloatCmd)
}

func (r *RedisClient) GeoHash(ctx context.Context, key string, members ...string) *redis.StringSliceCmd {
	args := r.Called(ctx, key, members)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (r *RedisClient) ACLDryRun(ctx context.Context, username string, command ...any) *redis.StringCmd {
	args := r.Called(ctx, username, command)
	return args.Get(0).(*redis.StringCmd)
}

func (r *RedisClient) AddHook(hook redis.Hook) {
	r.Called(hook)
}

func (r *RedisClient) Watch(ctx context.Context, fn func(*redis.Tx) error, keys ...string) error {
	args := r.Called(ctx, fn, keys)
	return args.Error(0)
}

func (r *RedisClient) Do(ctx context.Context, args ...any) *redis.Cmd {
	call := r.Called(ctx, args)
	return call.Get(0).(*redis.Cmd)
}

func (r *RedisClient) Process(ctx context.Context, cmd redis.Cmder) error {
	args := r.Called(ctx, cmd)
	return args.Error(0)
}

func (r *RedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	args := r.Called(ctx, channels)
	return args.Get(0).(*redis.PubSub)
}

func (r *RedisClient) PSubscribe(ctx context.Context, channels ...string) *redis.PubSub {
	args := r.Called(ctx, channels)
	return args.Get(0).(*redis.PubSub)
}

func (r *RedisClient) SSubscribe(ctx context.Context, channels ...string) *redis.PubSub {
	args := r.Called(ctx, channels)
	return args.Get(0).(*redis.PubSub)
}

func (r *RedisClient) Close() error {
	args := r.Called()
	return args.Error(0)
}

func (r *RedisClient) PoolStats() *redis.PoolStats {
	args := r.Called()
	return args.Get(0).(*redis.PoolStats)
}
