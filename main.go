package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/0x6flab/namegenerator"
	"gopkg.in/ini.v1"
)

type Config struct {
	TotalPoints       int     `json:"total_points"`
	NumArms           int     `json:"num_arms"`
	PitchAngleDegrees float64 `json:"pitch_angle_degrees"`
	MinDistance       float64 `json:"min_distance"`
	RadiusLimit       float64 `json:"radius_limit"`
	GalaxyType        string  `json:"galaxy_type"`
	WormholesRaw      string  `json:"wormholes_raw"`
	NumPlayers        int     `json:"num_players"`
	MaxDistToNeighbor float64 `json:"max_dist_to_neighbor"`
	NonPlayerMinR     int     `json:"non_player_min_r"`
	NonPlayerMaxR     int     `json:"non_player_max_r"`
	NonPlayerMinG     int     `json:"non_player_min_g"`
	NonPlayerMaxG     int     `json:"non_player_max_g"`
	NonPlayerMinE     int     `json:"non_player_min_e"`
	NonPlayerMaxE     int     `json:"non_player_max_e"`
	NonPlayerMinI     int     `json:"non_player_min_i"`
	NonPlayerMaxI     int     `json:"non_player_max_i"`
	NonPlayerMinS     int     `json:"non_player_min_s"`
	NonPlayerMaxS     int     `json:"non_player_max_s"`
	InitialR          int     `json:"r"`
	InitialG          int     `json:"g"`
	InitialE          int     `json:"e"`
	InitialI          int     `json:"i"`
	InitialS          int     `json:"s"`
	InitialST         int     `json:"st"`
}

type Star struct {
	UID  int     `json:"uid"`
	Name string  `json:"name"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	R    int     `json:"r"`
	G    int     `json:"g"`
	E    int     `json:"e"`
	I    int     `json:"i"`
	S    int     `json:"s"`
	ST   int     `json:"st"`
	PUID *int    `json:"puid,omitempty"`
}

type MapData struct {
	Stars     []Star  `json:"stars"`
	Wormholes [][]int `json:"wormholes"`
}

type PageData struct {
	Config   Config
	JSONText string
	Message  string
	Error    string
}

var (
	pageTmpl *template.Template
	rng      = rand.New(rand.NewSource(time.Now().UnixNano()))
	nameGen  = namegenerator.NewGenerator()
	randMu   sync.Mutex
	nameMu   sync.Mutex
)

func main() {
	var err error
	pageTmpl, err = template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatalf("failed to load template: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/generate", handleGenerate)

	addr := ":8080"
	log.Printf("neptune-map-go-htmx listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	cfg := loadConfigFromINI("config.ini")
	jsonText, err := generateMapJSON(cfg)
	if err != nil {
		renderPage(w, PageData{Config: cfg, Error: err.Error()})
		return
	}
	renderPage(w, PageData{Config: cfg, JSONText: jsonText, Message: "Loaded default config and generated JSON."})
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		renderJSONPanel(w, "", "invalid form submission")
		return
	}

	cfg, err := configFromForm(r)
	if err != nil {
		renderJSONPanel(w, "", err.Error())
		return
	}

	jsonText, err := generateMapJSON(cfg)
	if err != nil {
		renderJSONPanel(w, "", err.Error())
		return
	}

	renderJSONPanel(w, jsonText, "")
}

func renderPage(w http.ResponseWriter, data PageData) {
	if err := pageTmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderJSONPanel(w http.ResponseWriter, jsonText, errMsg string) {
	const panel = `<div id="json-panel">{{if .Err}}<div class="error">{{.Err}}</div>{{end}}<textarea id="jsonText" name="json_text">{{.JSON}}</textarea><div class="button-row"><button type="button" class="secondary" onclick="copyJson()">Copy JSON</button><button type="button" onclick="refreshMap()">Refresh Map From JSON</button></div></div>`
	t := template.Must(template.New("panel").Parse(panel))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = t.Execute(w, map[string]string{"JSON": jsonText, "Err": errMsg})
}

func loadConfigFromINI(path string) Config {
	cfg := Config{
		TotalPoints:       600,
		NumArms:           12,
		PitchAngleDegrees: 25,
		MinDistance:       1.5,
		RadiusLimit:       50,
		GalaxyType:        "spiral",
		WormholesRaw:      "[[179,425],[216,440],[315,446],[367,127],[473,432]]",
		NumPlayers:        10,
		MaxDistToNeighbor: 100,
		NonPlayerMinR:     0,
		NonPlayerMaxR:     50,
		NonPlayerMinG:     0,
		NonPlayerMaxG:     1,
		NonPlayerMinE:     0,
		NonPlayerMaxE:     0,
		NonPlayerMinI:     0,
		NonPlayerMaxI:     0,
		NonPlayerMinS:     0,
		NonPlayerMaxS:     0,
		InitialR:          49,
		InitialG:          0,
		InitialE:          200,
		InitialI:          600,
		InitialS:          500,
		InitialST:         300,
	}

	if _, err := os.Stat(path); err != nil {
		return cfg
	}

	f, err := ini.Load(path)
	if err != nil {
		return cfg
	}

	cfg.TotalPoints = f.Section("galaxy").Key("total_points").MustInt(cfg.TotalPoints)
	cfg.NumArms = f.Section("galaxy").Key("num_arms").MustInt(cfg.NumArms)
	cfg.PitchAngleDegrees = f.Section("galaxy").Key("pitch_angle_degrees").MustFloat64(cfg.PitchAngleDegrees)
	cfg.MinDistance = f.Section("galaxy").Key("min_distance").MustFloat64(cfg.MinDistance)
	cfg.RadiusLimit = f.Section("galaxy").Key("radius_limit").MustFloat64(cfg.RadiusLimit)
	cfg.GalaxyType = strings.TrimSpace(strings.ToLower(f.Section("galaxy").Key("galaxy_type").MustString(cfg.GalaxyType)))
	cfg.WormholesRaw = strings.TrimSpace(f.Section("wormholes").Key("pairs").MustString(cfg.WormholesRaw))
	cfg.NumPlayers = f.Section("player_positioning").Key("num_players").MustInt(cfg.NumPlayers)
	cfg.MaxDistToNeighbor = f.Section("player_positioning").Key("max_dist_to_neighbor").MustFloat64(cfg.MaxDistToNeighbor)
	cfg.NonPlayerMinR = f.Section("non_player_resource").Key("min_r").MustInt(cfg.NonPlayerMinR)
	cfg.NonPlayerMaxR = f.Section("non_player_resource").Key("max_r").MustInt(cfg.NonPlayerMaxR)
	cfg.NonPlayerMinG = f.Section("non_player_resource").Key("min_g").MustInt(cfg.NonPlayerMinG)
	cfg.NonPlayerMaxG = f.Section("non_player_resource").Key("max_g").MustInt(cfg.NonPlayerMaxG)
	cfg.NonPlayerMinE = f.Section("non_player_resource").Key("min_e").MustInt(cfg.NonPlayerMinE)
	cfg.NonPlayerMaxE = f.Section("non_player_resource").Key("max_e").MustInt(cfg.NonPlayerMaxE)
	cfg.NonPlayerMinI = f.Section("non_player_resource").Key("min_i").MustInt(cfg.NonPlayerMinI)
	cfg.NonPlayerMaxI = f.Section("non_player_resource").Key("max_i").MustInt(cfg.NonPlayerMaxI)
	cfg.NonPlayerMinS = f.Section("non_player_resource").Key("min_s").MustInt(cfg.NonPlayerMinS)
	cfg.NonPlayerMaxS = f.Section("non_player_resource").Key("max_s").MustInt(cfg.NonPlayerMaxS)
	cfg.InitialR = f.Section("player_initial_resource").Key("r").MustInt(cfg.InitialR)
	cfg.InitialG = f.Section("player_initial_resource").Key("g").MustInt(cfg.InitialG)
	cfg.InitialE = f.Section("player_initial_resource").Key("e").MustInt(cfg.InitialE)
	cfg.InitialI = f.Section("player_initial_resource").Key("i").MustInt(cfg.InitialI)
	cfg.InitialS = f.Section("player_initial_resource").Key("s").MustInt(cfg.InitialS)
	cfg.InitialST = f.Section("player_initial_resource").Key("st").MustInt(cfg.InitialST)

	return cfg
}

func configFromForm(r *http.Request) (Config, error) {
	cfg := Config{}
	var err error

	cfg.TotalPoints, err = parseIntField(r, "total_points")
	if err != nil {
		return cfg, err
	}
	cfg.NumArms, err = parseIntField(r, "num_arms")
	if err != nil {
		return cfg, err
	}
	cfg.PitchAngleDegrees, err = parseFloatField(r, "pitch_angle_degrees")
	if err != nil {
		return cfg, err
	}
	cfg.MinDistance, err = parseFloatField(r, "min_distance")
	if err != nil {
		return cfg, err
	}
	cfg.RadiusLimit, err = parseFloatField(r, "radius_limit")
	if err != nil {
		return cfg, err
	}
	cfg.GalaxyType = strings.ToLower(strings.TrimSpace(r.FormValue("galaxy_type")))
	cfg.WormholesRaw = strings.TrimSpace(r.FormValue("wormholes_raw"))
	cfg.NumPlayers, err = parseIntField(r, "num_players")
	if err != nil {
		return cfg, err
	}
	cfg.MaxDistToNeighbor, err = parseFloatField(r, "max_dist_to_neighbor")
	if err != nil {
		return cfg, err
	}
	cfg.NonPlayerMinR, err = parseIntField(r, "non_player_min_r")
	if err != nil {
		return cfg, err
	}
	cfg.NonPlayerMaxR, err = parseIntField(r, "non_player_max_r")
	if err != nil {
		return cfg, err
	}
	cfg.NonPlayerMinG, err = parseIntField(r, "non_player_min_g")
	if err != nil {
		return cfg, err
	}
	cfg.NonPlayerMaxG, err = parseIntField(r, "non_player_max_g")
	if err != nil {
		return cfg, err
	}
	cfg.NonPlayerMinE, err = parseIntField(r, "non_player_min_e")
	if err != nil {
		return cfg, err
	}
	cfg.NonPlayerMaxE, err = parseIntField(r, "non_player_max_e")
	if err != nil {
		return cfg, err
	}
	cfg.NonPlayerMinI, err = parseIntField(r, "non_player_min_i")
	if err != nil {
		return cfg, err
	}
	cfg.NonPlayerMaxI, err = parseIntField(r, "non_player_max_i")
	if err != nil {
		return cfg, err
	}
	cfg.NonPlayerMinS, err = parseIntField(r, "non_player_min_s")
	if err != nil {
		return cfg, err
	}
	cfg.NonPlayerMaxS, err = parseIntField(r, "non_player_max_s")
	if err != nil {
		return cfg, err
	}
	cfg.InitialR, err = parseIntField(r, "r")
	if err != nil {
		return cfg, err
	}
	cfg.InitialG, err = parseIntField(r, "g")
	if err != nil {
		return cfg, err
	}
	cfg.InitialE, err = parseIntField(r, "e")
	if err != nil {
		return cfg, err
	}
	cfg.InitialI, err = parseIntField(r, "i")
	if err != nil {
		return cfg, err
	}
	cfg.InitialS, err = parseIntField(r, "s")
	if err != nil {
		return cfg, err
	}
	cfg.InitialST, err = parseIntField(r, "st")
	if err != nil {
		return cfg, err
	}

	if cfg.TotalPoints <= 0 {
		return cfg, fmt.Errorf("total_points must be > 0")
	}
	if cfg.GalaxyType != "spiral" && cfg.GalaxyType != "elliptical" && cfg.GalaxyType != "irregular" {
		return cfg, fmt.Errorf("galaxy_type must be spiral, elliptical, or irregular")
	}
	if cfg.NonPlayerMinR < 0 || cfg.NonPlayerMinG < 0 || cfg.NonPlayerMinE < 0 || cfg.NonPlayerMinI < 0 || cfg.NonPlayerMinS < 0 {
		return cfg, fmt.Errorf("non-player min resources must be >= 0")
	}
	if cfg.NonPlayerMaxR < cfg.NonPlayerMinR || cfg.NonPlayerMaxG < cfg.NonPlayerMinG || cfg.NonPlayerMaxE < cfg.NonPlayerMinE || cfg.NonPlayerMaxI < cfg.NonPlayerMinI || cfg.NonPlayerMaxS < cfg.NonPlayerMinS {
		return cfg, fmt.Errorf("non-player max resources must be >= min resources")
	}

	return cfg, nil
}

func parseIntField(r *http.Request, key string) (int, error) {
	v := strings.TrimSpace(r.FormValue(key))
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid %s", key)
	}
	return n, nil
}

func parseFloatField(r *http.Request, key string) (float64, error) {
	v := strings.TrimSpace(r.FormValue(key))
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s", key)
	}
	return n, nil
}

func generateMapJSON(cfg Config) (string, error) {
	wormholes := make([][]int, 0)
	if err := json.Unmarshal([]byte(cfg.WormholesRaw), &wormholes); err != nil {
		return "", fmt.Errorf("wormholes must be valid JSON array of pairs")
	}

	coords, err := generateCoordinates(cfg)
	if err != nil {
		return "", err
	}

	stars := make([]Star, 0, len(coords))
	for i, p := range coords {
		starName := generateStarName()
		if starName == "" {
			starName = fmt.Sprintf("Star-%03d", i+1)
		}
		r, g, e, ind, s := nonPlayerStarResources(cfg)
		star := Star{
			UID:  i + 1,
			Name: starName,
			X:    round3(p[0]),
			Y:    round3(p[1]),
			R:    r,
			G:    g,
			E:    e,
			I:    ind,
			S:    s,
			ST:   0,
		}
		stars = append(stars, star)
	}

	stars, err = assignPlayers(stars, cfg)
	if err != nil {
		return "", err
	}

	result := MapData{Stars: stars, Wormholes: wormholes}
	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func generateCoordinates(cfg Config) ([][2]float64, error) {
	coords := make([][2]float64, 0, cfg.TotalPoints)
	maxAttempts := cfg.TotalPoints * 10000
	pitch := cfg.PitchAngleDegrees * math.Pi / 180

	for attempts := 0; len(coords) < cfg.TotalPoints && attempts < maxAttempts; attempts++ {
		var x, y float64
		switch cfg.GalaxyType {
		case "spiral":
			radius := randFloat64() * cfg.RadiusLimit
			angle := randFloat64() * 2 * math.Pi
			armOffset := radius / math.Tan(pitch)
			armAngle := angle + armOffset
			arm := randIntn(maxInt(1, cfg.NumArms))
			armRotation := 2 * math.Pi * float64(arm) / float64(maxInt(1, cfg.NumArms))
			armAngle += armRotation
			x = radius * math.Cos(armAngle)
			y = radius * math.Sin(armAngle)
		case "elliptical":
			radius := randFloat64() * cfg.RadiusLimit
			angle := randFloat64() * 2 * math.Pi
			x = radius * math.Cos(angle)
			y = radius * math.Sin(angle)
		case "irregular":
			x = (randFloat64()*2 - 1) * cfg.RadiusLimit
			y = (randFloat64()*2 - 1) * cfg.RadiusLimit
		default:
			return nil, fmt.Errorf("unknown galaxy type")
		}

		ok := true
		for _, c := range coords {
			if math.Hypot(x-c[0], y-c[1]) < cfg.MinDistance {
				ok = false
				break
			}
		}
		if ok {
			coords = append(coords, [2]float64{x, y})
		}
	}

	if len(coords) < cfg.TotalPoints {
		return nil, fmt.Errorf("could not place %d stars; try reducing min_distance or total_points", cfg.TotalPoints)
	}
	return coords, nil
}

func assignPlayers(stars []Star, cfg Config) ([]Star, error) {
	if cfg.NumPlayers <= 0 {
		return stars, nil
	}
	if cfg.NumPlayers > len(stars) {
		return nil, fmt.Errorf("num_players cannot exceed total stars")
	}

	indices := randPerm(len(stars))[:cfg.NumPlayers]
	optimizedMin := cfg.RadiusLimit / (2 * math.Sqrt(float64(cfg.NumPlayers)))
	optimizedMax := cfg.RadiusLimit / (2 * math.Sqrt(math.Log(float64(maxInt(2, cfg.NumPlayers)))))
	optimizedDist := cfg.MaxDistToNeighbor
	if optimizedDist < optimizedMin || optimizedDist > optimizedMax {
		optimizedDist = (optimizedMin + optimizedMax) / 2
	}

	for i := 1; i < len(indices); i++ {
		tries := 0
		for {
			ok := true
			for j := 0; j < i; j++ {
				a := stars[indices[i]]
				b := stars[indices[j]]
				if math.Hypot(a.X-b.X, a.Y-b.Y) <= optimizedDist {
					ok = false
					break
				}
			}
			if ok {
				break
			}
			indices[i] = randIntn(len(stars))
			tries++
			if tries > len(stars)*10 {
				break
			}
		}
	}

	for i, idx := range indices {
		p := i + 1
		stars[idx].PUID = &p
		stars[idx].R = cfg.InitialR
		stars[idx].G = cfg.InitialG
		stars[idx].E = cfg.InitialE
		stars[idx].I = cfg.InitialI
		stars[idx].S = cfg.InitialS
		stars[idx].ST = cfg.InitialST
	}

	return stars, nil
}

func generateStarName() string {
	nameMu.Lock()
	defer nameMu.Unlock()
	return nameGen.Generate()
}

func nonPlayerStarResources(cfg Config) (int, int, int, int, int) {
	return randomResourceRange(cfg.NonPlayerMinR, cfg.NonPlayerMaxR), randomResourceRange(cfg.NonPlayerMinG, cfg.NonPlayerMaxG), randomResourceRange(cfg.NonPlayerMinE, cfg.NonPlayerMaxE), randomResourceRange(cfg.NonPlayerMinI, cfg.NonPlayerMaxI), randomResourceRange(cfg.NonPlayerMinS, cfg.NonPlayerMaxS)
}

func randomResource(max int) int {
	if max <= 0 {
		return 0
	}
	return randIntn(max + 1)
}

func randomResourceRange(min, max int) int {
	if max <= min {
		return min
	}
	return min + randIntn(max-min+1)
}

func randFloat64() float64 {
	randMu.Lock()
	defer randMu.Unlock()
	return rng.Float64()
}

func randIntn(n int) int {
	randMu.Lock()
	defer randMu.Unlock()
	return rng.Intn(n)
}

func randPerm(n int) []int {
	randMu.Lock()
	defer randMu.Unlock()
	return rng.Perm(n)
}

func round3(v float64) float64 {
	return math.Round(v*1000) / 1000
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
