# 版本号分布式管理系统

# 一、项目功能模块介绍
## 1. Raft 节点管理模块
#### 该模块使用 github.com/hashicorp/raft 库实现 Raft 一致性算法，负责管理分布式系统中的节点。主要功能包括节点的初始化、集群的构建和节点的加入。
#### 节点初始化时，会设置节点的 ID、快照间隔和阈值等配置信息，同时初始化日志存储和状态存储。
#### 集群构建时，会创建 Raft 节点，并根据节点的角色（领导者、跟随者等）执行相应的操作。
#### 节点加入功能允许新节点加入现有的 Raft 集群，确保集群的扩展性。
## 2. 数据库操作模块
#### 该模块使用 gorm.io/gorm 和 github.com/go-sql-driver/mysql 库实现与 MySQL 数据库的交互。主要功能包括数据库的连接、表的创建和数据的增删改查。
#### 通过 GORM 的 ORM 功能，可以将数据库表映射为 Go 语言的结构体，方便进行数据操作。
## 3. 缓存管理模块
#### 该模块使用 github.com/go-redis/redis/v8 库实现与 Redis 数据库的交互。主要功能包括缓存的设置、获取和删除。
#### 缓存管理模块可以提高数据的访问速度，减轻数据库的压力

# 二、结构体介绍 version.go
## 结构体
### 1.Version 结构体
#### Version 结构体用于表示版本信息，它与数据库中的版本信息表相对应。
#### 包含三个字段：
#### ①ID：版本的唯一标识符，使用 gorm:"primaryKey" 标签表示该字段是数据库表的主键。
#### ②VersionNo：版本号，gorm:"not null" 标签表示该字段在数据库中不能为空。
#### ③Platform：版本适用的平台，同样使用 gorm:"not null" 标签表示不能为空。
### 2.VersionDB结构体
#### VersionDB 结构体也用于表示版本信息，它与 Version 结构体的功能类似，但增加了一些额外的标签，主要用于 JSON 序列化和数据验证。
#### 包含三个字段：
#### ①ID：版本记录的唯一标识符，是主键。json:"id" 标签表示该字段在 JSON 序列化时的字段名为 "id"，validate:"required" 标签表示该字段是必填项。
#### ②VersionNo：版本号，json:"versionNo" 标签表示该字段在 JSON 序列化时的字段名为 "versionNo"，validate:"required" 标签表示该字段是必填项。
#### ③Platform：平台名称，json:"platform" 标签表示该字段在 JSON 序列化时的字段名为 "platform"，validate:"required" 标签表示该字段是必填项。

# 三、数据库初始化
### 1.mysql.go 负责对 mysql 数据库进行连接
### 2.cache.go 负责对 redis 数据库进行连接

# 四、数据库交互
### 1.version_mysql_dao.go 中的 VersionMysqlDao 封装了与 MySQL 数据库的连接，通过 DB 字段与 MySQL 数据库进行交互，提供了一系列操作版本信息的方法。
### 2.version_redis_dao.go 中的 VersionRedisDao 封装了与 Redis 数据库的连接，通过 Client 字段与 Redis 数据库进行交互，提供了一系列操作版本信息的方法。

# 五、服务端代码
## service 包主要负责处理与版本号相关的业务逻辑，通过调用不同的数据访问层（Redis 和 MySQL）来实现对版本信息的增删改查操作，并结合 Raft 集群来保证数据的一致性和可靠性。
## 1. version_redis_service.go 实现了与 Redis 相关的版本号服务
## 主要功能
### ①VersionRedisService 结构体：封装了对 Redis 数据操作的 redisDao。
### ②NewVersionRedisService 函数：用于创建 VersionRedisService 实例。
### ③VersionExists 方法：判断 Redis 中指定 ID 的版本是否存在。
### ④AddVersion 方法：向 Redis 中添加版本信息。
### ⑤DeleteVersion 方法：从 Redis 中删除指定 ID 的版本信息。
### ⑥UpdateVersion 方法：更新 Redis 中指定 ID 的版本信息。
### ⑦GetVersionByID 方法：从 Redis 中获取指定 ID 的版本信息
## 2. version_mysql_service.go 实现了与 Mysql 相关的版本号服务
## 主要功能
### ①VersionMysqlService 结构体：封装了对 MySQL 数据操作的 mysqlDao。
### ②NewVersionMysqlService 函数：用于创建 VersionMysqlService 实例，同时会检查 mysqlDao 是否为 nil。
### ③ConvertToVersion 方法：将从数据库获取的版本信息转换为服务使用的版本格式。
### ④VersionExists 方法：判断 MySQL 中指定 ID 的版本是否存在。
### ⑤AddVersionToMysql 方法：向 MySQL 中添加版本信息。
### ⑥GetVersionFromMysql 方法：从 MySQL 中获取指定 ID 的版本信息。
### ⑦UpdateVersion 方法：更新 MySQL 中指定 ID 的版本信息，使用事务进行操作。
### ⑧DeleteVersion 方法：从 MySQL 中删除指定 ID 的版本信息。
### ⑨GetAllVersions 方法：获取 MySQL 中所有的版本信息。
## 3. version_service.go 实现了综合的版本号服务，整合了 MySQL 和 Redis 的操作，并包含一些与 Raft 相关的操作。
## 主要功能
### ①VersionService 结构体：整合了 VersionMysqlService、VersionRedisService 和 Raft 节点信息。
### ②NewVersionService 函数：用于创建 VersionService 实例，同时会初始化 Raft 节点。
### ③其他方法：包含了一系列与版本号操作相关的方法，如添加版本、更新版本、删除版本等，部分方法涉及到 Raft 集群的操作。

# 六、控制层代码
## controller 包的主要作用是作为 Web 应用的控制器层，接收和处理来自客户端的 HTTP 请求，调用服务层的方法处理业务逻辑，并将处理结果返回给客户端，实现了版本号管理和节点加入 Raft 集群的功能。
## 主要功能
### ①VersionController 结构体：包含一个指向 service.VersionService 的指针，用于调用服务层的方法处理版本号相关业务。
### ②NewVersionController 函数：用于创建 VersionController 实例。
### ③AddVersion 方法：处理添加版本号信息的请求，接收客户端发送的 JSON 格式的版本号信息，调用服务层的 AddVersion 方法添加版本号，根据处理结果返回相应的 HTTP 状态码和消息。
### ④DeleteVersion 方法：处理删除版本号信息的请求，从 URL 参数中获取版本号 ID，调用服务层的 DeleteVersion 方法删除版本号，根据处理结果返回相应的 HTTP 状态码和消息。
### ⑤UpdateVersion 方法：处理更新版本号信息的请求，从 URL 参数中获取版本号 ID，检查版本号是否存在，接收客户端发送的 JSON 格式的更新后的版本号信息，调用服务层的 UpdateVersion 方法更新版本号，根据处理结果返回相应的 HTTP 状态码和消息。
### ⑥GetVersionByID 方法：处理获取指定 ID 的版本号信息的请求，从路由参数中获取版本号 ID，调用服务层的 GetVersionByID 方法获取版本号信息，根据处理结果返回相应的 HTTP 状态码和版本号信息。
### ⑦JoinRaftCluster 方法：处理将节点加入 Raft 集群的请求，从请求路径参数中获取节点 ID 和地址，调用服务层的 JoinRaftCluster 方法尝试加入 Raft 集群，根据处理结果返回相应的 HTTP 状态码和消息。

# 七、路由代码
## routers 包的主要作用是定义和设置 HTTP 路由，将不同的 URL 路径映射到相应的控制器方法上，实现了版本管理和 Raft 集群相关请求的分发和处理。通过调用 SetUpVersionRouter 和 SetUpRaftRouter 函数，可以分别设置版本管理和 Raft 相关的路由。
## 1.SetUpVersionRouter ：
### 创建一个默认的 Gin 引擎实例，定义了与版本管理相关的路由，包括添加、删除、更新、获取版本号信息以及节点加入 Raft 集群的路由，并将这些路由映射到 VersionController 的相应方法上，最后返回该 Gin 引擎实例。
## 2.SetUpRaftRouter ：
### 创建一个默认的 Gin 引擎实例，定义了一个根路径的 GET 请求，用于处理节点加入 Raft 集群的请求，将该路由映射到 VersionController 的 JoinRaftCluster 方法上，最后返回该 Gin 引擎实例。

# 八、main.go
## 1.startNode 
### 负责启动单个节点，包括数据库和缓存的初始化、数据访问对象（DAO）的创建、服务的初始化、控制器的创建、路由的设置以及服务器的启动。如果任何一步出现错误，程序将记录错误日志并终止。
## 2.main 
### 作为程序的入口点，首先获取配置信息，然后初始化 Redis 和 MySQL 数据库。
### 创建 MySQL 和 Redis 的数据访问对象和服务。
### 启动一个定期同步任务，将 MySQL 数据同步到 Redis。
### 遍历所有 Raft 节点，使用 sync.WaitGroup 并发启动每个节点。
## 3.startPeriodicSync
### 启动一个定期任务，每隔一段时间将 MySQL 数据同步到 Redis。使用 time.Ticker 来实现定时功能，调用版本服务的 ReLoadCacheDataInternal 方法进行数据同步，并记录日志。

# 九、配置文件
## 定义应用程序的配置结构体以及提供获取配置实例的方法。
## 1.Config 结构体
### 该结构体是整个应用程序配置的核心，包含了多个子结构体，分别用于存储不同组件的配置信息：
### ①MySQL：存储 MySQL 数据库的配置信息，主要是 DSN，用于连接 MySQL 数据库。
### ②Redis：存储 Redis 数据库的配置信息，包括地址、密码和数据库编号。
### ③MemoryDB：存储内存数据库的配置信息，如容量和淘汰比例。
### ④Server：存储服务器相关的配置信息，包括重载缓存数据的时间间隔、定期删除过期键的时间间隔和检测数量。
### ⑤Raft：存储 Raft 集群的节点信息，是一个 Node 结构体的切片。
## 2.Node 结构体
### 该结构体用于表示 Raft 集群中的一个节点，包含节点的 ID、Address 和 Port 信息。
## 3.GetConfig 函数
### 该函数用于获取应用程序的配置实例，返回一个 Config 结构体。函数内部返回一个包含默认配置信息的 Config 结构体实例，这些默认配置信息可以在需要时进行修改。

# 十、raft实现多实例部署的数据一致性
## raft 包通过 RaftInitializerImpl 初始化 Raft 节点，使用 VersionFSM 处理 Raft 节点的状态变化，通过 NewRaftNode 方法创建和启动 Raft 节点，实现了 Raft 一致性算法在项目中的应用。
## 1.RaftInitializerImpl
### 实现了 Raft 节点的初始化器，通过 InitRaft 方法来初始化一个 Raft 节点。
## 2.InitRaft 
### 创建一个 VersionFSM 有限状态机实例，用于处理 Raft 节点的状态变化。
### 调用 node.NewRaftNode 方法创建一个新的 Raft 节点，并将其返回
## 3.VersionFSM
### 实现了 Raft 有限状态机，包含一个 VersionServiceInterface 服务实例，用于处理具体的业务逻辑。
## 4.Apply 
### 根据 Raft 日志中的命令类型，调用相应的服务方法进行处理。
## 5.Snapshot 
### 实现了快照逻辑，当前返回 nil，表示没有实现具体的快照功能。
## 6.NewRaftNode 
### 初始化 Raft 配置，包括节点 ID、快照间隔和阈值等。
### 初始化日志存储、状态存储和快照存储。
### 初始化传输层。
### 创建 Raft 节点。
### 如果是第一个节点，初始化集群；否则，尝试加入现有集群。

# 接口文档截图
## 添加版本号信息(AddVersion)
<img width="1279" src="https://github.com/user-attachments/assets/38eab37f-f477-45e4-8b00-e54d874b0752" /><br>
## 更新版本号信息(UpdateVersion)
<img width="1279" src="https://github.com/user-attachments/assets/fd70115d-cb70-43ec-9f28-68a8a94932e5" /><br>
## 删除版本号信息(DeleteVersion)
<img width="1279" src="https://github.com/user-attachments/assets/e4a2eddf-3d6b-44e8-a158-199c6fac4509" /><br>
## 查询版本号信息(GetVersionByID)
<img width="1279" src="https://github.com/user-attachments/assets/f70275bc-2f0d-4a0e-96ec-21ecc75d3432" /><br>



