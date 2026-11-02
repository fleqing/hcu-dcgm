# topology-info 使用说明

获取整机 HCU 的互联拓扑矩阵，并以 JSON 输出到标准输出。

## 功能

调用 `dcgm.DiscoverInterconnectTopology()`，枚举所有 HCU 两两之间的互联关系，返回 `HCU × HCU` 矩阵。每个单元格描述“源设备 → 目标设备”的链路信息。

## 运行环境

- 已安装 HCU 驱动及 rsmi 库，能够正常枚举到 HCU 设备
- Go 环境，且 cgo 可用（`pkg/dcgm` 依赖 cgo，需要 `drm/drm.h` 等头文件，仅能在 Linux 目标机上编译）
- 具有访问 HCU 设备及 `/dev/kfd` 的权限

## 运行

在项目根目录执行：

```bash
go run ./samples/topology-info
```

或先编译再运行：

```bash
go build -o topology-info ./samples/topology-info
./topology-info
```

### 命令行参数

程序调用了 `flag.Parse()`，支持 glog 的标准参数：

| 参数 | 说明 |
| --- | --- |
| `-logtostderr` | 日志输出到标准错误，而不是写入日志文件 |
| `-v=<N>` | 日志详细级别，`-v=5` 可输出链路状态、远端 BDF 等调试信息 |
| `-vmodule=<pattern>=<N>` | 按文件设置日志级别，例如 `-vmodule=api=5` |

JSON 结果输出到标准输出，日志输出到标准错误，两者互不干扰。

### 调试示例

```bash
# 输出 JSON，同时打印详细日志
go run ./samples/topology-info -logtostderr -v=5

# 只保留 XHCL 链路相关日志，并过滤掉 JSON
go run ./samples/topology-info -logtostderr -v=5 2>&1 >/dev/null | grep -E "xhcl|XHCL"

# 只保存 JSON 结果，便于后续用 jq 处理
go run ./samples/topology-info > topo.json
jq '.Matrix[4][5]' topo.json
```

## 输出格式

```json
{
  "DeviceCount": 8,
  "Matrix": [
    [
      {
        "SrcDvInd": 0,
        "DstDvInd": 1,
        "BdfID": 11520,
        "PciID": "0000:2d:00.0",
        "LinkType": "HSL",
        "Weight": 3,
        "Hops": 0
      }
    ]
  ]
}
```

顶层字段：

| 字段 | 说明 |
| --- | --- |
| `DeviceCount` | HCU 设备数量 |
| `Matrix` | `DeviceCount × DeviceCount` 的二维数组，`Matrix[src][dst]` 表示 src 到 dst 的链路 |

单元格字段：

| 字段 | 说明 |
| --- | --- |
| `SrcDvInd` | 源设备索引 |
| `DstDvInd` | 目标设备索引 |
| `BdfID` | 目标设备的 BDF ID（十进制数值） |
| `PciID` | 目标设备的 PCI 地址，格式 `dddd:bb:dd.f` |
| `LinkType` | 链路类型，见下表 |
| `Weight` | 链路权重，计算规则见下表 |
| `Hops` | 跳数，目前固定为 0 |

### LinkType 取值

| 取值 | 含义 |
| --- | --- |
| `NONE` | 无链路，或源与目标为同一设备 |
| `PCIE` | 仅通过 PCIe 互联 |
| `HSL` | 通过 HSL 直连（非 Hyswitch） |
| `HYSWITCH` | 通过 Hyswitch 互联 |
| `HYBRID` | PCIe 与 HSL 混合互联 |
| `UNKNOWN` | 无法识别的链路类型 |

### Weight 计算规则

| 场景 | Weight |
| --- | --- |
| 源与目标为同一设备 | `-1` |
| `PCIE`，两设备在同一 NUMA 节点 | `1` |
| `PCIE`，两设备在不同 NUMA 节点 | `0` |
| `HSL` | 源设备上 **状态为 UP** 且 **远端 BDF 等于目标设备 BDF** 的 XHCL 链路条数 |
| `HYSWITCH` | `DeviceCount - 1` |

## 结果解读

以 8 卡、两块板（设备 0–3 与 4–7）为例，期望得到：

- 同一板内的设备之间为 `HSL`，`Weight` 为直连的 XHCL 链路条数
- 跨板的设备之间为 `PCIE`
- 对角线为 `NONE`，`Weight` 为 `-1`

## 常见问题

### 某些设备间 `LinkType` 为 `HSL`，但 `Weight` 为 0

`LinkType` 由 `rsmi_topo_get_link_type` 判定，`Weight` 则依赖 XHCL 链路状态和远端 BDF。出现 `HSL` 且 `Weight` 为 0，说明类型识别成功，但没有统计到符合条件的链路，可能原因：

1. 该设备的 XHCL 链路状态不是 UP，或状态查询失败
2. 读取远端 BDF（`rsmi_dev_xhcl_link_remote_bdfid_get`）失败
3. 驱动返回的远端 BDF 与 `rsmiDevPciIdGet` 的格式不一致，导致比较不相等

排查方法：使用 `-logtostderr -v=5` 运行，查看对应设备的以下日志：

- `rsmi_dev_xhcl_link_number_get`：链路数量
- `XHCL link x/y state=... up=...`：每条链路的状态
- `remote bdfid = 0x...`：远端 BDF 的值
- `get remote bdfid failed`：读取失败的告警

### 编译报错 `drm/drm.h file not found`

`pkg/dcgm` 通过 cgo 引用了 Linux 内核的 DRM 头文件，在 macOS 上无法编译。请在安装了 HCU 驱动和 libdrm 开发包的 Linux 机器上编译运行。
