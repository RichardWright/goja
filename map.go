package goja

type mapEntry struct {
	key, value Value

	iterPrev, iterNext *mapEntry
	hNext              *mapEntry
}

type orderedMap struct {
	hash                map[uint64]*mapEntry
	iterFirst, iterLast *mapEntry
	size                int
}

type orderedMapIter struct {
	m   *orderedMap
	cur *mapEntry
}

func (m *orderedMap) lookup(key Value) (h uint64, entry, hPrev *mapEntry) {
	if key == _negativeZero {
		key = intToValue(0)
	}
	h = key.hash()
	for entry = m.hash[h]; entry != nil && !entry.key.SameAs(key); hPrev, entry = entry, entry.hNext {
	}
	return
}

func (m *orderedMap) set(key, value Value) {
	h, entry, hPrev := m.lookup(key)
	if entry != nil {
		entry.value = value
	} else {
		if key == _negativeZero {
			key = intToValue(0)
		}
		entry = &mapEntry{key: key, value: value}
		if hPrev == nil {
			m.hash[h] = entry
		} else {
			hPrev.hNext = entry
		}
		if m.iterLast != nil {
			entry.iterPrev = m.iterLast
			m.iterLast.iterNext = entry
		} else {
			m.iterFirst = entry
		}
		m.iterLast = entry
		m.size++
	}
}

func (m *orderedMap) get(key Value) Value {
	_, entry, _ := m.lookup(key)
	if entry != nil {
		return entry.value
	}

	return nil
}

func (m *orderedMap) remove(key Value) bool {
	h, entry, hPrev := m.lookup(key)
	if entry != nil {
		entry.key = nil
		entry.value = nil

		// remove from the doubly-linked list
		if entry.iterPrev != nil {
			entry.iterPrev.iterNext = entry.iterNext
		} else {
			m.iterFirst = entry.iterNext
		}
		if entry.iterNext != nil {
			entry.iterNext.iterPrev = entry.iterPrev
		} else {
			m.iterLast = entry.iterPrev
		}

		// remove from the hash
		if hPrev == nil {
			if entry.hNext == nil {
				delete(m.hash, h)
			} else {
				m.hash[h] = entry.hNext
			}
		} else {
			hPrev.hNext = entry.hNext
		}

		m.size--
		return true
	}

	return false
}

func (m *orderedMap) has(key Value) bool {
	_, entry, _ := m.lookup(key)
	return entry != nil
}

func (iter *orderedMapIter) next() *mapEntry {
	if iter.m == nil {
		// closed iterator
		return nil
	}
	cur := iter.cur
	if cur != nil {
		cur = cur.iterNext
		// skip deleted entries
		for cur != nil && cur.key == nil {
			cur = cur.iterNext
		}
		iter.cur = cur
	} else {
		iter.cur = iter.m.iterFirst
	}
	if iter.cur == nil {
		iter.close()
	}
	return iter.cur
}

func (iter *orderedMapIter) close() {
	iter.m = nil
}

func newOrderedMap() *orderedMap {
	return &orderedMap{
		hash: make(map[uint64]*mapEntry),
	}
}

func (m *orderedMap) newIter() *orderedMapIter {
	iter := &orderedMapIter{
		m: m,
	}
	return iter
}

func (m *orderedMap) clear() {
	for item := m.iterFirst; item != nil; item = item.iterNext {
		item.key = nil
		item.value = nil
		if item.iterPrev != nil {
			item.iterPrev.iterNext = nil
		}
	}
	m.iterFirst = nil
	m.iterLast = nil
	m.hash = make(map[uint64]*mapEntry)
	m.size = 0
}