module github.com/rinx/vald-meta-halodb

go 1.14

replace (
	github.com/cockroachdb/errors => github.com/cockroachdb/errors v1.5.0
	github.com/vdaas/vald/apis => github.com/vdaas/vald/apis v0.0.44
)

require (
	cloud.google.com/go v0.60.0
	code.cloudfoundry.org/bytefmt v0.0.0-20200131002437-cf55d5288a48
	contrib.go.opencensus.io/exporter/jaeger v0.2.0
	contrib.go.opencensus.io/exporter/prometheus v0.2.0
	contrib.go.opencensus.io/exporter/stackdriver v0.13.2
	github.com/aws/aws-sdk-go v1.33.5
	github.com/cockroachdb/errors v0.0.0-00010101000000-000000000000
	github.com/danielvladco/go-proto-gql/pb v0.6.1 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-redis/redis/v7 v7.4.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gocql/gocql v0.0.0-20200624222514-34081eda590e
	github.com/gocraft/dbr/v2 v2.7.0
	github.com/gogo/protobuf v1.3.1
	github.com/gorilla/mux v1.7.4
	github.com/hashicorp/go-version v1.2.1
	github.com/json-iterator/go v1.1.10
	github.com/klauspost/compress v1.10.10
	github.com/kpango/fastime v1.0.16
	github.com/kpango/gache v1.2.1
	github.com/kpango/glg v1.5.1
	github.com/pierrec/lz4/v3 v3.3.2
	github.com/scylladb/gocqlx v1.5.0
	github.com/tensorflow/tensorflow v2.2.0+incompatible
	github.com/vdaas/vald v0.0.45
	go.opencensus.io v0.22.4
	go.uber.org/automaxprocs v1.3.0
	go.uber.org/goleak v1.0.0 // indirect
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	golang.org/x/sys v0.0.0-20200625212154-ddb9806d33ae
	google.golang.org/api v0.29.0
	google.golang.org/grpc v1.30.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.5
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/metrics v0.18.5
	sigs.k8s.io/controller-runtime v0.6.1
)
