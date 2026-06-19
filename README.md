# Neptune Map Generator (Go + HTMX)

A modern, interactive, and responsive web application designed to generate and visualize custom galaxy maps for space strategy games (such as _Neptune's Pride_).

Built with a high-performance **Go** backend for coordinate generation and mathematical modeling, **HTMX** for smooth asynchronous page updates, and **Plotly.js** for rich, interactive galaxy visualizations.

---

## 🚀 Features

- **Dynamic Galaxy Geometry**:
  - **Spiral**: Multi-armed spiral galaxies with custom arm counts and pitch angles.
  - **Elliptical**: Evenly distributed oval/circular systems.
  - **Irregular**: Scattered, chaotic star clustering.
- **Resource Distribution Logic**:
  - Customize range boundaries (`min`/`max`) for economy (`e`), industry (`i`), science (`s`), scanning range (`st`), resources (`r`), and warp gates (`g`) on non-player stars.
  - Control starting resource baselines for player home systems.
- **Optimized Player Spacing**:
  - Distribute player starting systems evenly across the map with configurable minimum and maximum neighborhood distances.
- **Interactive Visualization**:
  - Zoom, pan, and hover over stars to see detailed resource breakdowns, IDs, names, and owner stats.
  - Interactive wormhole links highlighted across coordinate pairs.
- **Seamless Live-Editing**:
  - Modify parameters and hit **Generate JSON** to compute new coordinates.
  - Live-edit the JSON text directly in the browser and hit **Refresh Map From JSON** to instantly re-plot your changes.
  - Copy JSON output to your clipboard with a single click.
- **Configuration Persistence**:
  - Fallback and default configuration parameters are read directly from `config.ini`.

---

## 🛠️ Architecture

- **`main.go`**: Contains the core logic, including:
  - Configuration parsing (from `config.ini` and HTTP forms).
  - Galaxy math (spiral arm generation using logarithmic spiral approximations, boundary checks, and collision detection to maintain `min_distance`).
  - Asynchronous player spacing optimization.
  - Star name generator integration (`github.com/0x6flab/namegenerator`).
- **`config.ini`**: Initial defaults for total stars, player count, wormhole arrays, and resource spreads.
- **`templates/index.html`**: The UI template styled with a custom dark space-age theme, featuring HTMX form submissions and Plotly map initialization/drawing functions.

---

## 📋 Prerequisites

To run this project locally, you need:

- **Go** (version 1.22 or higher)

---

## 🏃 Getting Started

1. **Clone or Download the Project**:
   Ensure you are inside the project root directory:

   ```bash
   cd neptune-map-go-htmx
   ```

2. **Install Dependencies**:
   Download the required Go modules:

   ```bash
   go mod tidy
   ```

3. **Run the Application**:
   Start the Go web server:

   ```bash
   go run main.go
   ```

4. **Access the Web Interface**:
   Open your browser and navigate to:
   [http://localhost:8080](http://localhost:8080)

---

## ⚙️ Configuration (`config.ini`)

The project reads baseline configurations from `config.ini`. An example structure is shown below:

```ini
[galaxy]
total_points = 600
num_arms = 12
pitch_angle_degrees = 25.0
min_distance = 1.5
radius_limit = 50.0
galaxy_type = spiral

[wormholes]
# JSON array containing pairs of Star UIDs to connect via wormholes
pairs = [[179,425],[216,440],[315,446],[367,127],[473,432]]

[player_positioning]
num_players = 10
max_dist_to_neighbor = 100

[non_player_resource]
min_r = 0
max_r = 50
min_g = 0
max_g = 1
min_e = 0
max_e = 0
min_i = 0
max_i = 0
min_s = 0
max_s = 0

[player_initial_resource]
r = 49
g = 0
e = 200
i = 600
s = 500
st = 300
```

---

## 📄 Output Data Schema

The generated JSON output consists of two main fields: `stars` and `wormholes`.

### Star Object

Each star has the following parameters:

- `uid` _(int)_: Unique star identifier (1-indexed).
- `name` _(string)_: Randomly generated name.
- `x` / `y` _(float64)_: Coordinates (rounded to 3 decimal places).
- `r` _(int)_: Resource value.
- `g` _(int)_: Gate presence (1 if gate exists, 0 otherwise).
- `e` _(int)_: Economy.
- `i` _(int)_: Industry.
- `s` _(int)_: Science.
- `st` _(int)_: Scanning range or other custom game-state attributes.
- `puid` _(int, optional)_: Player owner UID (only present if the star is a player home world).

### Wormholes List

An array of coordinate pairs:

```json
"wormholes": [
  [179, 425],
  [216, 440]
]
```

---

## 🎨 Technologies Used

- **Go**: Fast backend JSON processing and galaxy math.
- **HTMX**: High-performance hypermedia-driven UI components.
- **Plotly.js**: Dynamic interactive plotting and graphing library.
- **gopkg.in/ini.v1**: INI file reader for default configuration parameters.
- **Name Generator**: Go name generator module for unique celestial labels.

## Deployment

Deployment can be found in [here](./docs/deployment.md)

## Licensing

[MIT License](./LICENSE)
