package job

import (
	"bytes"
	"image"
	"image/png"
	"log"
	"math"
	"math/rand"
	"strings"

	"github.com/oldfritter/lucy/dom"
	"github.com/oldfritter/lucy/internal/cache"
	"github.com/oldfritter/lucy/internal/pool"
	captchaImage "github.com/oldfritter/lucy/lib/captcha"
	"github.com/oldfritter/lucy/lib/db"
	"github.com/oldfritter/lucy/lib/storage/oss"
	"github.com/oldfritter/lucy/model"
)

// 默认系统词库 — 常用成语（每个成语内无重复字）
var systemWordBank = []string{
	"一马当先", "一鸣惊人", "一箭双雕", "万众一心", "三心二意", "三言两语",
	"五光十色", "五湖四海", "七上八下", "七拼八凑", "九牛一毛", "九死一生",
	"大公无私", "大器晚成", "大材小用", "山清水秀", "山穷水尽", "山盟海誓",
	"千山万水", "千军万马", "千锤百炼", "马到成功", "心花怒放", "心想事成",
	"心平气和", "心旷神怡", "天长地久", "天高云淡", "天罗地网", "风和日丽",
	"风调雨顺", "风雨同舟", "风花雪月", "风平浪静", "龙飞凤舞", "龙争虎斗",
	"龙马精神", "龙腾虎跃", "虎头蛇尾", "虎背熊腰", "虎踞龙盘", "对牛弹琴",
	"鸟语花香", "画蛇添足", "画龙点睛", "叶公好龙", "守株待兔", "掩耳盗铃",
	"亡羊补牢", "百折不挠", "百尺竿头", "百感交集", "百川归海", "更进一步",
	"金碧辉煌", "金玉良言", "金枝玉叶", "金戈铁马", "金口玉言", "金榜题名",
	"花好月圆", "花言巧语", "花团锦簇", "花枝招展", "春暖花开", "春华秋实",
	"春回大地", "春风化雨", "秋高气爽", "秋毫无犯", "秋收冬藏", "冰天雪地",
	"冰清玉洁", "水滴石穿", "水落石出", "水到渠成", "水涨船高", "水深火热",
	"火冒三丈", "火上浇油", "火中取栗", "火树银花", "光明正大", "光宗耀祖",
	"光怪陆离", "光彩夺目", "日新月异", "日积月累", "日久天长", "日月如梭",
	"月明星稀", "月下花前", "星罗棋布", "星火燎原", "星移斗转", "星驰电走",
	"地大物博", "地久天长", "地动山摇", "地灵人杰", "万紫千红", "万象更新",
	"万无一失", "万马奔腾", "人声鼎沸", "人面桃花", "人定胜天", "手舞足蹈",
	"手忙脚乱", "手到擒来", "手疾眼快", "眉飞色舞", "眉清目秀", "眉开眼笑",
	"目瞪口呆", "目不转睛", "目不暇接", "目光如炬", "目中无人", "目无全牛",
	"口是心非", "口若悬河", "口蜜腹剑", "口出狂言", "口诛笔伐", "耳闻目睹",
	"耳目一新", "耳聪目明", "耳提面命", "耳熟能详", "头重脚轻", "头破血流",
	"头昏眼花", "脚踏实地", "脚忙手乱", "胸有成竹", "胸无大志", "胸罗万象",
	"胸襟开阔", "心惊肉跳", "心照不宣", "心血来潮", "自以为是", "自强不息",
	"自食其力", "自告奋勇", "精卫填海", "愚公移山", "女娲补天", "开天辟地",
	"夸父追日", "后羿射日", "嫦娥奔月", "牛郎织女", "八仙过海", "叶落归根",
	"根深蒂固", "瓜熟蒂落", "落花流水", "流连忘返", "返璞归真", "真理名言",

	// 五字短语（text5 专用，字不重复）
	"今晚打老虎", "明天去游泳", "明天会更好", "春风吹又生",
	"更上一层楼", "床前明月光", "低头思故乡", "锄禾日当午",
	"春眠不觉晓", "花落知多少", "家和万事兴", "民以食为天",
	"日久见人心", "家书抵万金", "名师出高徒", "功到自然成",
	"学而时习之", "温故而知新", "四海皆兄弟", "相识满天下",
	"礼轻情意重", "人穷志不短", "天涯若比邻", "海内存知己",
	"少壮不努力", "老大徒伤悲", "欲速则不达", "鲤鱼跳龙门",
	"岁月不待人", "读书破万卷", "下笔如有神", "路遥知马力",
	"疾风知劲草", "烈火见真金", "岁寒知松柏", "时穷节乃见",
	"人生如朝露", "光阴似箭飞", "日月如穿梭", "白驹过隙间",
	"一日难再晨", "盛年不重来",

	// 六字短语（text6 专用，字不重复）
	"有志者事竟成", "天有不测风云", "人有旦夕祸福", "远水不救近火",
	"真金不怕火炼", "英雄所见略同", "二者不可得兼", "万变不离其宗",
	"三人行有我师", "比上不足下余", "杀鸡焉用牛刀", "车到山前有路",
	"风马牛不相及", "春风又绿江南", "海阔凭鱼跃天", "莫愁前路无知",
	"挂羊头卖狗肉", "近朱者赤墨黑", "海水不可斗量", "满招损谦受益",
	"牛头不对马嘴", "不可同日而语", "秋水共长天色", "百思不得其解",
	"三寸不烂之舌", "耳闻不如目见", "百闻不如一见", "闻名不如见面",
	"船到桥头自然", "落霞与孤鹜飞", "有过之无不及", "顾左右而言他",
	"过五关斩六将", "迅雷不及掩耳", "事实胜于雄辩",
}

const (
	batchPerRun = 20 // 每轮每个投放最多生成 20 个验证码
)

func init() {
	Register(Job{
		Name: "fill-campaign",
		Spec: "@every 1m",
		Func: fillCampaignCaptchas,
	})
}

// fillCampaignCaptchas 扫描未足额的投放并补充生成验证码
func fillCampaignCaptchas() {
	var campaigns []model.Campaign
	if err := db.MysqlDB.Where("status IN (1)").Find(&campaigns).Error; err != nil {
		log.Printf("[fill-campaign] query campaigns failed: %v", err)
		return
	}

	if len(campaigns) == 0 {
		return
	}

	log.Printf("[fill-campaign] scanning %d active campaign(s)", len(campaigns))

	for _, campaign := range campaigns {
		// 统计已生成数量
		count := countCaptchasByCampaign(campaign.Id)

		needed := campaign.CaptchaCount - int(count)
		if needed <= 0 {
			// 用户投放达上限 → 标记完成；系统投放跳过本轮（池中有足够待用验证码）
			if campaign.Type == dom.CampaignTypeUser {
				markCompleted(&campaign)
			}
			continue
		}

		// 标记为处理中
		if campaign.Status == 0 {
			db.MysqlDB.Model(&campaign).Update("status", 1)
		}

		// 每轮最多生成 batchPerRun 个
		toGenerate := needed
		if toGenerate > batchPerRun {
			toGenerate = batchPerRun
		}

		log.Printf("[fill-campaign] campaign=%d name=%s deficit=%d generating=%d",
			campaign.Id, campaign.Name, needed, toGenerate)

		created := 0
		for i := 0; i < toGenerate; i++ {
			if err := generateOne(&campaign); err != nil {
				log.Printf("[fill-campaign] campaign=%d generate failed: %v", campaign.Id, err)
				continue
			}
			created++
		}

		log.Printf("[fill-campaign] campaign=%d created=%d", campaign.Id, created)

		// 重新统计
		count = countCaptchasByCampaign(campaign.Id)
		if int(count) >= campaign.CaptchaCount {
			if campaign.Type == dom.CampaignTypeUser {
				markCompleted(&campaign)
			}
		}
	}
}

func markCompleted(campaign *model.Campaign) {
	db.MysqlDB.Model(campaign).Update("status", 2)
	log.Printf("[fill-campaign] campaign=%d completed", campaign.Id)
}

// generateOne 为投放创建一条验证码记录并生成图片
func generateOne(campaign *model.Campaign) error {
	prompts := pickPrompts(campaign)
	if len(prompts) == 0 {
		prompts = pickSysPrompts(campaign.CaptchaType)
	}

	switch campaign.CaptchaType {
	case "rotate":
		return createRotateCaptcha(campaign)
	case "text5":
		return createText5Captcha(campaign, prompts)
	case "text6":
		return createText6Captcha(campaign, prompts)
	default: // text4
		return createText4Captcha(campaign, prompts)
	}
}

// pickPrompts 从词库中随机选取一个与验证码类型匹配的词组，按顺序返回各字
// 词库格式：好好学习，天天向上，叶公好龙，武松打虎（text4）
//
//	不到长城非好汉，春蚕到死丝方尽（text7）
func pickPrompts(campaign *model.Campaign) []string {
	if campaign.WordBank == "" {
		return nil
	}
	want := promptCount(campaign.CaptchaType)
	phrases := strings.FieldsFunc(campaign.WordBank, func(r rune) bool {
		return r == '，' || r == ','
	})
	var valid []string
	for _, p := range phrases {
		p = strings.TrimSpace(p)
		if len([]rune(p)) == want {
			valid = append(valid, p)
		}
	}
	if len(valid) == 0 {
		return nil
	}
	// 随机选一个匹配长度的词组
	word := valid[rand.Intn(len(valid))]
	result := make([]string, want)
	for i, r := range word {
		result[i] = string(r)
	}
	return result
}

// pickSysPrompts 从系统成语库按长度匹配词组；无匹配时取任意一个成语
func pickSysPrompts(captchaType string) []string {
	n := promptCount(captchaType)

	// 收集匹配长度的成语
	var matched []string
	for _, idiom := range systemWordBank {
		if len([]rune(idiom)) == n {
			matched = append(matched, idiom)
		}
	}
	if len(matched) == 0 {
		// 无匹配长度（如 text5/text6 无对应成语），取任意成语并截断
		matched = systemWordBank
	}

	word := []rune(matched[rand.Intn(len(matched))])
	if len(word) > n {
		word = word[:n]
	}
	result := make([]string, n)
	for i, r := range word {
		result[i] = string(r)
	}
	return result
}

func promptCount(captchaType string) int {
	switch captchaType {
	case "text5":
		return 5
	case "text6":
		return 6
	default:
		return 4
	}
}

func createText4Captcha(campaign *model.Campaign, prompts []string) error {
	return createTextCaptcha(campaign, prompts, 4)
}

func createText5Captcha(campaign *model.Campaign, prompts []string) error {
	return createTextCaptcha(campaign, prompts, 5)
}

func createText6Captcha(campaign *model.Campaign, prompts []string) error {
	return createTextCaptcha(campaign, prompts, 6)
}

// createTextCaptcha 通用文字验证码创建：字符随机散布在背景图上，位点由图片渲染确定
func createTextCaptcha(campaign *model.Campaign, prompts []string, count int) error {
	if len(prompts) < count {
		// 用系统字库补齐
		sys := pickSysPrompts(campaign.CaptchaType)
		for len(prompts) < count && len(sys) > 0 {
			prompts = append(prompts, sys[0])
			sys = sys[1:]
		}
	}
	if len(prompts) > count {
		prompts = prompts[:count]
	}

	// 生成挑战图 + 获取各字位点
	img, points := captchaImage.GenerateTextChallenge(prompts)

	// 先创建 DB 记录触发 BeforeCreate（生成 Key、Captcha）
	var c model.CaptchaText4
	switch count {
	case 5:
		cc := model.CaptchaText5{
			Prompt1: prompts[0], Prompt2: prompts[1], Prompt3: prompts[2], Prompt4: prompts[3], Prompt5: prompts[4],
		}
		cc.UserId = campaign.UserId
		cc.CampaignId = &campaign.Id
		tx := db.BeginTx()
		defer tx.DbRollback()
		if err := tx.Create(&cc).Error; err != nil {
			return err
		}
		setVerifyPositions5(&cc, points)
		if err := uploadAndCache(&cc, cc.Key, img); err != nil {
			return err
		}
		tx.Save(&cc)
		tx.DbCommit()
		pool.AddToPool("text5", cc.Uid)
		return nil
	case 6:
		cc := model.CaptchaText6{
			Prompt1: prompts[0], Prompt2: prompts[1], Prompt3: prompts[2], Prompt4: prompts[3], Prompt5: prompts[4], Prompt6: prompts[5],
		}
		cc.UserId = campaign.UserId
		cc.CampaignId = &campaign.Id
		tx := db.BeginTx()
		defer tx.DbRollback()
		if err := tx.Create(&cc).Error; err != nil {
			return err
		}
		setVerifyPositions6(&cc, points)
		if err := uploadAndCache(&cc, cc.Key, img); err != nil {
			return err
		}
		tx.Save(&cc)
		tx.DbCommit()
		pool.AddToPool("text6", cc.Uid)
		return nil
	default:
		c = model.CaptchaText4{
			Prompt1: prompts[0], Prompt2: prompts[1], Prompt3: prompts[2], Prompt4: prompts[3],
		}
	}
	c.UserId = campaign.UserId
	c.CampaignId = &campaign.Id

	tx := db.BeginTx()
	defer tx.DbRollback()
	if err := tx.Create(&c).Error; err != nil {
		return err
	}
	setVerifyPositions4(&c, points)
	if err := uploadAndCache(&c, c.Key, img); err != nil {
		return err
	}
	tx.Save(&c)
	tx.DbCommit()
	pool.AddToPool("text4", c.Uid)
	return nil
}

func setVerifyPositions4(c *model.CaptchaText4, points []image.Point) {
	if len(points) >= 4 {
		c.Verify1X, c.Verify1Y = points[0].X, points[0].Y
		c.Verify2X, c.Verify2Y = points[1].X, points[1].Y
		c.Verify3X, c.Verify3Y = points[2].X, points[2].Y
		c.Verify4X, c.Verify4Y = points[3].X, points[3].Y
	}
}

func setVerifyPositions5(c *model.CaptchaText5, points []image.Point) {
	if len(points) >= 5 {
		c.Verify1X, c.Verify1Y = points[0].X, points[0].Y
		c.Verify2X, c.Verify2Y = points[1].X, points[1].Y
		c.Verify3X, c.Verify3Y = points[2].X, points[2].Y
		c.Verify4X, c.Verify4Y = points[3].X, points[3].Y
		c.Verify5X, c.Verify5Y = points[4].X, points[4].Y
	}
}

func setVerifyPositions6(c *model.CaptchaText6, points []image.Point) {
	if len(points) >= 6 {
		c.Verify1X, c.Verify1Y = points[0].X, points[0].Y
		c.Verify2X, c.Verify2Y = points[1].X, points[1].Y
		c.Verify3X, c.Verify3Y = points[2].X, points[2].Y
		c.Verify4X, c.Verify4Y = points[3].X, points[3].Y
		c.Verify5X, c.Verify5Y = points[4].X, points[4].Y
		c.Verify6X, c.Verify6Y = points[5].X, points[5].Y
	}
}

// captchaRecord 文字验证码需暴露的字段
type captchaRecord interface {
	GetCaptcha() string
	Json() string
}

// uploadAndCache 将图片编码为 PNG 上传 OSS（用 Key 作路径），并写入 Redis 缓存
func uploadAndCache(c captchaRecord, ossKey string, img image.Image) error {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}
	b := buf.Bytes()
	if _, err := oss.PutObject(ossKey, &b); err != nil {
		return err
	}
	return cache.SetCaptchaCache(&scheduleCaptcha{cc: c, data: c.Json()})
}

// scheduleCaptcha 适配 cache.CaptchaCacher 接口
type scheduleCaptcha struct {
	cc   captchaRecord
	data string
}

func (sc *scheduleCaptcha) GetCaptcha() string { return sc.cc.GetCaptcha() }
func (sc *scheduleCaptcha) Json() string       { return sc.data }

func createRotateCaptcha(campaign *model.Campaign) error {
	c := model.CaptchaImageRotate{
		Indicator: "▲",
		Tolerance: 15,
	}
	c.UserId = campaign.UserId
	c.CampaignId = &campaign.Id

	c.Angle = float64(randInt(30, 330))
	// 让角度偏离竖直方向一定程度
	c.Angle = math.Mod(c.Angle+180, 360) - 180 // [-180, 180]

	tx := db.BeginTx()
	defer tx.DbRollback()
	if err := tx.Create(&c).Error; err != nil {
		return err
	}
	c.Create()
	tx.DbCommit()
	pool.AddToPool("rotate", c.Uid)
	return nil
}

func randInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

// countCaptchasByCampaign 跨四张验证码表统计指定投放已生成的验证码总数
func countCaptchasByCampaign(campaignId int) int64 {
	var total int64
	tables := []any{
		&model.CaptchaText4{},
		&model.CaptchaText5{},
		&model.CaptchaText6{},
		&model.CaptchaImageRotate{},
	}
	for _, t := range tables {
		var n int64
		db.MysqlDB.Model(t).Where("campaign_id = ?", campaignId).Count(&n)
		total += n
	}
	return total
}
