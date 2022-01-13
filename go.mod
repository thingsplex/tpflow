module github.com/thingsplex/tpflow

require (
	github.com/ChrisTrenkamp/goxpath v0.0.0-20170922090931-c385f95c6022
	github.com/Knetic/govaluate v3.0.0+incompatible
	github.com/asdine/storm v2.1.1+incompatible
	github.com/boltdb/bolt v1.3.1
	github.com/cpucycle/astrotime v0.0.0-20120927164819-9c7d514efdb5
	github.com/dchest/uniuri v0.0.0-20160212164326-8902c56451e9
	github.com/futurehomeno/fimpgo v1.7.0
	github.com/google/uuid v1.2.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/imdario/mergo v0.3.6
	github.com/influxdata/influxdb v1.6.2
	github.com/kelvins/sunrisesunset v0.0.0-20170601204625-14f1915ad4b4
	github.com/labstack/gommon v0.2.1
	github.com/mitchellh/mapstructure v1.0.0
	github.com/nathan-osman/go-sunrise v0.0.0-20171121204956-7c449e7c690b
	github.com/oliveagle/jsonpath v0.0.0-20180606110733-2e52cf6e6852
	github.com/pkg/errors v0.8.0
	github.com/robfig/cron/v3 v3.0.0
	github.com/sirupsen/logrus v1.8.1
	github.com/thingsplex/tprelay v0.0.0-20210623220309-5ffe8cbfac9b
	github.com/traefik/yaegi v0.9.16
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)

require (
	github.com/DataDog/zstd v1.3.4 // indirect
	github.com/Sereal/Sereal v0.0.0-20180727013122-68c42fd7bfdf // indirect
	github.com/buger/jsonparser v1.0.0 // indirect
	github.com/coreos/bbolt v1.3.0 // indirect
	github.com/eclipse/paho.mqtt.golang v1.2.0 // indirect
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/valyala/bytebufferpool v0.0.0-20160817181652-e746df99fe4a // indirect
	github.com/valyala/fasttemplate v0.0.0-20170224212429-dcecefd839c4 // indirect
	github.com/vmihailenco/msgpack v4.0.0+incompatible // indirect
	golang.org/x/net v0.0.0-20180826012351-8a410e7b638d // indirect
	golang.org/x/sys v0.0.0-20191026070338-33540a1f6037 // indirect
	golang.org/x/text v0.3.0 // indirect
	google.golang.org/appengine v1.1.0 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
)

// remove after testing
//replace github.com/thingsplex/tprelay => ../tprelay

go 1.17
