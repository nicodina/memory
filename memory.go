package memory

import (
	"container/list"
	"sync"
	"time"
	"github.com/nicodina/session"
)

var pder = &ProviderMemory{list: list.New()}

// SessionStore implements Session
type SessionStore struct {
	sid          string                      // unique session id
	timeAccessed time.Time                   // last access time
	value        map[interface{}]interface{} // session value stored inside
}

// Set stores data in a session
func (st *SessionStore) Set(key, value interface{}) error {
	st.value[key] = value
	pder.SessionUpdate(st.sid)
	return nil
}

// Get retrieves data from a session
func (st *SessionStore) Get(key interface{}) interface{} {
	pder.SessionUpdate(st.sid)
	if v, ok := st.value[key]; ok {
		return v
	}
	return nil

}

// Delete removes data from a session
func (st *SessionStore) Delete(key interface{}) error {
	delete(st.value, key)
	pder.SessionUpdate(st.sid)
	return nil
}

// SessionID retrieves the session id
func (st *SessionStore) SessionID() string {
	return st.sid
}

// ProviderMemory implements Provider
type ProviderMemory struct {
	lock     sync.Mutex               // lock
	sessions map[string]*list.Element // save in memory
	list     *list.List               // gc
}

// SessionInit initializes a session
func (pder *ProviderMemory) SessionInit(sid string) (session.Session, error) {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	v := make(map[interface{}]interface{}, 0)
	newsess := &SessionStore{sid: sid, timeAccessed: time.Now(), value: v}
	element := pder.list.PushBack(newsess)
	pder.sessions[sid] = element
	return newsess, nil
}

// SessionRead retrieves a specific session
func (pder *ProviderMemory) SessionRead(sid string) (session.Session, error) {
	if element, ok := pder.sessions[sid]; ok {
		return element.Value.(*SessionStore), nil
	}
	sess, err := pder.SessionInit(sid)
	return sess, err

}

// SessionDestroy removes a session
func (pder *ProviderMemory) SessionDestroy(sid string) error {
	if element, ok := pder.sessions[sid]; ok {
		delete(pder.sessions, sid)
		pder.list.Remove(element)
	}
	return nil
}

// SessionGC implements the garbage collection
func (pder *ProviderMemory) SessionGC(maxlifetime int64) {
	pder.lock.Lock()
	defer pder.lock.Unlock()

	for {
		element := pder.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*SessionStore).timeAccessed.Unix() + maxlifetime) < time.Now().Unix() {
			pder.list.Remove(element)
			delete(pder.sessions, element.Value.(*SessionStore).sid)
		} else {
			break
		}
	}
}

// SessionUpdate modifies an existing session
func (pder *ProviderMemory) SessionUpdate(sid string) error {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	if element, ok := pder.sessions[sid]; ok {
		element.Value.(*SessionStore).timeAccessed = time.Now()
		pder.list.MoveToFront(element)
		return nil
	}
	return nil
}

func init() {
	pder.sessions = make(map[string]*list.Element, 0)
	session.Register("memory", pder)
}
