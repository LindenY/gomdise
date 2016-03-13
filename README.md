#gomdies

gomdies is a simple orm alike library aiming to provide complex data structure store/fetch services for redis

##Usage

###Initialize
    Create a gomdies instance by calling gomdies.New(pool redis.Pool)

###Save data
    To save data, invoke gomdies.Save(arg interface{}).
    1. Any exported fields of a struct will be saved by gomdies.
    2. To use the customized redis key for the data argument, implement mdl.Model interface


###Find data
    To find data, invoke gomdies.Find(key string, arg interface{})
    1. The arg parameter should be an pointer value, more actually, an addressable and settable parameter


###Sample usage
    For the example of how to use gomdise, see the (test file)[https://github.com/LindenY/gomdise/blob/v0.0.1-SNAPSHOT/gomdise_test.go] for more details.
