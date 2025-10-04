[![Go Report Card](https://goreportcard.com/badge/github.com/hx-w/minidemo-encoder)](https://goreportcard.com/report/github.com/hx-w/minidemo-encoder)
# Mini Demo Encoder

此工具用于解析CS:GO Demo文件(.dem)中玩家数据，并输出 [**BotMimic**](https://github.com/peace-maker/botmimic) 可读的bot录制文件(.rec)

## Notice

由于Demo文件记录的数据与source引擎的用户接口数据并不完全吻合，故录像中的某些玩家行为不能完全还原。

该问题也许可以通过某种技术手段或某种算法解决，但是由于本人精力有限，仅以此源码为开端，尽量记录些实现细节，以达到抛砖引玉的目的。

## Usage

1. 拉取本仓库
   ```bash
   git clone https://github.com/csgowiki/minidemo-encoder.git
   ```
2. 下载需要解析的demo文件
   > 完美竞技平台的demo文件修改过格式，解析会出问题，建议用HLTV的demo
3. 安装golang环境

### 单文件解析

运行脚本解析单个demo文件：
```bash
go run cmd/main.go -file {demo_path}
```
`{demo_path}`为需要解析的demo文件路径

### 批量解析（推荐）

**方法一：使用快捷脚本**

Windows系统直接双击运行 `batch_parse.bat` 或在命令行中执行：
```batch
batch_parse.bat [demo文件夹路径]
```

Linux/Mac系统：
```bash
chmod +x batch_parse.sh
./batch_parse.sh [demo文件夹路径]
```

如果不指定路径，默认解析 `./demo` 文件夹中的所有 `.dem` 文件。

**方法二：使用命令行**

先编译：
```bash
# Windows
go build -o minidemo-encoder.exe .\cmd\main.go

# Linux/Mac
go build -o minidemo-encoder ./cmd/main.go
```

再执行批量解析：
```bash
# Windows
minidemo-encoder.exe -dir="path\to\demo\folder"

# Linux/Mac
./minidemo-encoder -dir="path/to/demo/folder"
```

### 输出结构

**单文件解析：** 解析后的玩家录像文件会以回合数为子文件夹，保存在当前目录的`output/`文件夹下。

**批量解析：** 每个demo的解析结果会保存在独立的子文件夹中，文件夹以demo名称命名。例如：
```
output/
├── match1/
│   ├── round1_T5-CT5/
│   │   ├── t/
│   │   └── ct/
│   └── round2_T4-CT5/
├── match2/
│   └── ...
└── tournament_final/
    └── ...
```

更多详细说明请查看 [批量解析说明.md](批量解析说明.md)。

## REC 文件平衡工具

批量解析后，某些地图的某些回合可能出现玩家数量不足的情况。REC 平衡工具可以自动补全，确保每个 t/ct 文件夹都有 **5 个不同的 .rec 文件**。

### 使用方法

**Windows:**
```batch
rec_balance.bat "D:\SteamLibrary\steamapps\common\Counter-Strike Global Offensive\csgo\addons\sourcemod\data\botmimic\demotest"
```

**Linux/Mac:**
```bash
chmod +x rec_balance.sh
./rec_balance.sh /path/to/botmimic/demotest
```

### 补全规则

- ✅ 只从同一地图文件夹下寻找文件（如 `de_dust2` 只从其他 `de_dust2` 文件夹复制）
- ✅ t 文件夹只从其他 t 文件夹复制，ct 只从 ct 复制
- ✅ 可以从同一局的其他回合或不同局的回合复制，只要不是同一个文件夹即可
- ✅ 不会从同一文件夹内重复复制
- ✅ 自动处理文件名冲突

更多详细说明请查看 [REC平衡工具说明.md](REC平衡工具说明.md)。

## BotMimic

原版的botmimic使用的sourcemod环境落后，如果你没有一个可以运行指定.rec文件的插件，可以参考我的另一个插件：[**csgowiki-pack v1.4.4**](https://github.com/csgowiki/csgowiki-pack/tree/dev-1.4.4) 来自行修改。

v1.4.4一直没有发布，可以在仓库里手动下载编译版本，如果有余力的话，可以参考源码 [**replay.sp**](https://github.com/csgowiki/csgowiki-pack/blob/dev-1.4.4/scripting/csgowiki/minidemo/replay.sp) 实现自己的回放管理插件。

此外，在该仓库中还有一份我修改过的 [**botmimic_fix.sp**](https://github.com/csgowiki/csgowiki-pack/blob/dev-1.4.4/scripting/botmimic_fix.sp) 用来去除多余依赖，并且适应sourcemod更新版本，如有需要请自取。

## Advance

如果你对此项目感兴趣，想要进一步了解相关内容，不妨从.rec文件的格式看起。

.rec是二进制文件，用于存储玩家每一frame的数据，文件格式可以参考 [**test/test_encode.py**](test/test_encode.py)

以下是目前遇到的一些比较关键的问题：

**1. bot位置偏移**

在source引擎中，如果在每一回合都设置bot的具体坐标，那么bot的动作会变得很不流畅，腿部会抽搐。需要通过`OnPlayerRunCmd`方法设置bot当前帧的速度大小与方向，由引擎自行计算下一帧bot的运动路径。

但是由于Demo文件中只记录了当前帧的玩家速度，并没有记录当前帧玩家的所有交互信息，所以无法准确给出当前帧玩家速度的变化，所以导致了生成的回放文件，在回放过程中bot随着时间的增加越来越偏离实际路线的情况。

为了规避这种现象，只能使用BotMimic提供的关键帧标记：可以选择某一帧为关键帧，在关键帧同步bot与demo中玩家的**位置**/**朝向**/**速度**，以实现归零误差的效果。

但也正如开头所说，如果关键帧设置的频率太高，bot的移动会非常不流畅，如果频率过低，bot的运动误差会过大。为了优化该问题，我尽可能通过已知的数据预测当前帧的玩家速度变化，减少运动误差：[internal/parser/utils.go#L109-L151](https://github.com/csgowiki/minidemo-encoder/blob/0762925497d26f15c728c5f37a5fd720470d2186/internal/parser/utils.go#L109-L151)，但是效果并不明显。

**2. 回合开始时的异常**

目前录像回放在回合开始时经常出现bot位置的异常，所以不得不将回合前2000帧全部设为关键帧。

**3. 没有异常捕获导致的运行时错误处理**

由于sourcemod没有异常捕获机制，导致在回放录像时，一旦出现错误没办法及时处理，造成大量error log以至影响服务器性能。


**4. bot初始化与死亡处理**

由于游戏机制原因，增加或删减bot会影响到当前回合的状态，所以比较合理的做法是：在回放多个bot时，预生成多个bot，在bot死亡时也不要立马删除bot。在这样的前提下怎么做好资源分配，以及提高用户体验就是一个问题。


## Future

我的最初规划是借用 [**hltv-utility-api**](https://github.com/csgowiki/hltv-utility-api) 的思路，提供一个高性能的，实时更新的CSGO demo-to-rec的下载站，每日更新HLTV上的职业比赛录像，玩家可以在服务器中使用特定插件下载录像，通过bot回放的形式观看对局。

下载站尝试过Tencent Cloud COS，解决方案见：[**minidemo-hltv**](https://github.com/csgowiki/minidemo-hltv)。

希望这些想法对你有所帮助。
