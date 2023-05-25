module github.com/acs-dl/gitlab-module-svc

go 1.19

require (
	github.com/Masterminds/squirrel v1.5.3
	github.com/ThreeDotsLabs/watermill v1.2.0
	github.com/ThreeDotsLabs/watermill-amqp/v2 v2.0.7
	github.com/acs-dl/auth-svc v0.0.0-20230511083241-b6fde9efdf0d
	github.com/alecthomas/kingpin v2.2.6+incompatible
	github.com/fatih/structs v1.1.0
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/google/uuid v1.3.0
	github.com/hashicorp/vault/api v1.9.1
	github.com/pkg/errors v0.9.1
	github.com/rubenv/sql-migrate v1.3.0
	gitlab.com/distributed_lab/ape v1.7.1
	gitlab.com/distributed_lab/figure v2.1.2+incompatible
	gitlab.com/distributed_lab/kit v1.11.1
	gitlab.com/distributed_lab/logan v3.8.1+incompatible
	gitlab.com/distributed_lab/running v1.6.0
	gitlab.com/distributed_lab/urlval v3.0.0+incompatible
)

require (
	cloud.google.com/go/compute v1.12.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.1 // indirect
	cloud.google.com/go/iam v0.8.0 // indirect
	cloud.google.com/go/kms v1.8.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.3.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.2.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.1.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/azkeys v0.9.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/keyvault/internal v0.7.1 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v0.7.0 // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/aws/aws-sdk-go v1.36.29 // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/certifi/gocertifi v0.0.0-20200211180108-c7c1fbc02894 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/getsentry/raven-go v0.2.0 // indirect
	github.com/getsentry/sentry-go v0.7.0 // indirect
	github.com/go-gorp/gorp/v3 v3.0.2 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang-jwt/jwt/v4 v4.4.3 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/jsonapi v0.0.0-20200226002910-c8283f632fb7 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.0 // indirect
	github.com/googleapis/gax-go/v2 v2.7.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.6 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.6 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jmoiron/sqlx v1.2.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/lib/pq v1.10.0 // indirect
	github.com/lithammer/shortuuid/v3 v3.0.7 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/pelletier/go-toml v1.9.3 // indirect
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8 // indirect
	github.com/rabbitmq/amqp091-go v1.2.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.8.1 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	gitlab.com/distributed_lab/knox/knox-fork v0.0.0-20230209124153-2ee532a76d71 // indirect
	gitlab.com/distributed_lab/lorem v0.2.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/crypto v0.6.0 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/oauth2 v0.0.0-20221014153046-6fdb5e3db783 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	google.golang.org/api v0.103.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20221201164419-0e50fba7f41c // indirect
	google.golang.org/grpc v1.51.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
