// Package resouces services hold all data logic for preset suites domain
// like querying to database for preset suites
package resouces

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/faruqisan/resilia/engine/suites/services"
	"github.com/faruqisan/resilia/pkg/cache"
	"github.com/google/uuid"
)

type (
	// Engine struct define all resource required params
	// like db, redis, or another external service
	Engine struct {
		cache cache.Engine
	}
)

const (
	keySuites                    = "resilia_suites"
	keySuite                     = "resilia_suite:%s"
	keySuiteResources            = "resilia_suite_resources:%s"
	keySuiteResourcesCreatedHash = "resilia_suite_created_res_hs:%s"
	keySuiteCreatedResources     = "resilia_suite_created_res:%s:%s" // example: key: resilia_suite_created_res:1:deployment value : [redis-deployment, postgre-deployment]
	cacheExpire                  = time.Hour * 24                    // 24h expire
)

// New function return setuped resources engine
func New(cache cache.Engine) *Engine {
	return &Engine{
		cache: cache,
	}
}

// Create function store given suite to database returning id of model
func (e *Engine) Create(name string) (string, error) {
	var (
		id = uuid.New().String()
	)

	m := services.Model{
		ID:   id,
		Name: name,
	}

	key := fmt.Sprintf(keySuite, id)
	err := e.setSuiteToCache(key, m)
	if err != nil {
		return "", err
	}

	err = e.cache.SAdd(keySuites, key).Err()

	return id, err
}

// Find function return suite model from given id
func (e *Engine) Find(id string) (services.Model, error) {
	var m services.Model
	key := fmt.Sprintf(keySuite, id)
	str, err := e.cache.Get(key).Result()
	if err != nil {
		return m, err
	}

	err = json.Unmarshal([]byte(str), &m)
	if err != nil {
		return m, err
	}
	return m, nil
}

func (e *Engine) setSuiteToCache(key string, m services.Model) error {
	byteModel, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return e.cache.Set(key, string(byteModel), cacheExpire).Err()
}

// CreateResource function append file resource to suite
func (e *Engine) CreateResource(suiteID string, resource services.FileResource) error {

	key := fmt.Sprintf(keySuiteResources, suiteID)

	res, err := json.Marshal(resource)
	if err != nil {
		return nil
	}

	return e.cache.SAdd(key, res).Err()
}

// GetSuiteResources function return suite resources
func (e *Engine) GetSuiteResources(suiteID string) ([]services.FileResource, error) {
	var (
		key       = fmt.Sprintf(keySuiteResources, suiteID)
		resources []services.FileResource
	)

	rawResources, err := e.cache.SMembers(key).Result()
	if err != nil {
		return resources, nil
	}

	for _, rawResource := range rawResources {
		var r services.FileResource
		err = json.Unmarshal([]byte(rawResource), &r)
		if err != nil {
			return resources, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

// AppendCreatedResource function
func (e *Engine) AppendCreatedResource(suiteID, kind, resourceName string) error {
	// hset and sadd
	keyHash := fmt.Sprintf(keySuiteResourcesCreatedHash, suiteID)
	keyList := fmt.Sprintf(keySuiteCreatedResources, suiteID, kind)

	err := e.cache.HSet(keyHash, kind, keyList).Err()
	if err != nil {
		return err
	}

	return e.cache.SAdd(keyList, resourceName).Err()
}

// GetSuiteCreatedResource function return suite's created resource
func (e *Engine) GetSuiteCreatedResource(suiteID string) (map[services.KubeKind][]string, error) {

	var (
		cr      = make(map[services.KubeKind][]string)
		err     error
		keyHash = fmt.Sprintf(keySuiteResourcesCreatedHash, suiteID)
	)

	keyLists, err := e.cache.HGetAll(keyHash).Result()
	if err != nil {
		return cr, err
	}

	for kind, keyList := range keyLists {
		resourceName, err := e.cache.SMembers(keyList).Result()
		if err != nil {
			return cr, err
		}
		kk := services.KubeKind(kind)
		cr[kk] = append(cr[kk], resourceName...)
	}

	return cr, err

}

// Get function return all suites on database
func (e *Engine) Get() ([]services.Model, error) {
	return []services.Model{}, nil
}
