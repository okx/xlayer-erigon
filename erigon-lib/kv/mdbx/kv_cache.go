package mdbx

// CachedTx 实现了一个内存事务，所有读写操作都先在内存中进行。
type CachedTx struct {

	// writes 存放每个 bucket 内的写入操作（key->value）
	writes map[string]map[string][]byte
	// deletes 存放每个 bucket 内的删除标记（key->true）
	deletes map[string]map[string]bool

	// 底层数据库引用，可以用来创建实际的 mdbx 事务（这里假设 db 已经初始化）
	db *MdbxTx
}

// NewCachedTx 创建一个新的 CachedTx，暂时不启动底层事务
func NewCachedTx(db *MdbxTx) *CachedTx {
	return &CachedTx{
		db:      db,
		writes:  make(map[string]map[string][]byte),
		deletes: make(map[string]map[string]bool),
	}
}

// getCacheKey 这里简单使用 key 的字符串形式，实际场景可能需要更加健壮的构造方式
func getCacheKey(key []byte) string {
	return string(key)
}

func (ctx *CachedTx) SetDb(db *MdbxTx) {
	ctx.db = db
}

// Get 查询操作：先检查内存缓存，再查询底层数据
func (ctx *CachedTx) Get(bucket string, key []byte) ([]byte, error) {
	k := getCacheKey(key)
	// 如果该 key 在删除缓存中，则视为不存在
	if delMap, ok := ctx.deletes[bucket]; ok {
		if delMap[k] {
			return nil, nil
		}
	}
	// 如果在写缓存中找到了修改，则直接返回
	if bucketMap, ok := ctx.writes[bucket]; ok {
		if val, ok := bucketMap[k]; ok {
			return val, nil
		}
	}

	result, err := ctx.db.RealGetOne(bucket, key)
	if err != nil {
		return nil, err
	}
	// 这里将结果保存到变量中
	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)
	// 注意：如果底层返回 nil 表示不存在，则直接返回 nil
	ctx.putToCache(bucket, keyCopy, result)

	return result, err
}

// Put 操作：仅在内存缓存中保存写入操作
func (ctx *CachedTx) Put(bucket string, key, value []byte) error {
	if ctx.writes[bucket] == nil {
		ctx.writes[bucket] = make(map[string][]byte)
	}
	k := getCacheKey(key)
	// 写入缓存更新
	ctx.writes[bucket][k] = value
	// 如果之前标记为删除，则取消删除标记
	if ctx.deletes[bucket] != nil {
		delete(ctx.deletes[bucket], k)
	}
	return nil
}

// Delete 操作：仅在内存缓存中标记删除，同时从写缓存中移除
func (ctx *CachedTx) Delete(bucket string, key []byte) error {
	k := getCacheKey(key)
	if ctx.deletes[bucket] == nil {
		ctx.deletes[bucket] = make(map[string]bool)
	}
	ctx.deletes[bucket][k] = true
	// 同时删除写缓存中的对应项（如果存在）
	if ctx.writes[bucket] != nil {
		delete(ctx.writes[bucket], k)
	}
	return nil
}

// 辅助函数：将从底层读到的数据放入缓存中（用于保证在同一事务中后续读到最新数据）
func (ctx *CachedTx) putToCache(bucket string, key, value []byte) {
	if ctx.writes[bucket] == nil {
		ctx.writes[bucket] = make(map[string][]byte)
	}
	ctx.writes[bucket][getCacheKey(key)] = value
}

// Commit 阶段：将内存缓存中的所有操作一次性提交到底层数据库
func (ctx *CachedTx) Commit() error {
	// 先将所有写入操作提交
	for bucket, bucketMap := range ctx.writes {
		for k, v := range bucketMap {
			if err := ctx.db.Put(bucket, []byte(k), v); err != nil {
				ctx.db.Rollback()
				return err
			}
		}
	}
	// 再将删除操作提交
	for bucket, delMap := range ctx.deletes {
		for k := range delMap {
			if err := ctx.db.Delete(bucket, []byte(k)); err != nil {
				ctx.db.Rollback()
				return err
			}
		}
	}
	return nil
}

// Rollback 用于取消当前缓存中的所有操作
func (ctx *CachedTx) Rollback() {
	// 清空内存缓存
	ctx.writes = make(map[string]map[string][]byte)
	ctx.deletes = make(map[string]map[string]bool)
	if ctx.db != nil {
		ctx.db.Rollback()
	}
}
