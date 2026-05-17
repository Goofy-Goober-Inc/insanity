package main

import (
	"fmt"
	"math"
	"math/cmplx"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// ===================================================================
// ФУНДАМЕНТАЛЬНЫЕ ФИЗИЧЕСКИЕ КОНСТАНТЫ
// Все значения в единицах СИ, МэВ или планковских единицах
// Источники: CODATA 2018, PDG 2023, Planck Collaboration 2018
// ===================================================================

const (
	// --- Планковские единицы (граница применимости известной физики) ---
	PLANCK_LENGTH      = 1.616255e-35 // метр: квант пространства, петлевая квантовая гравитация
	PLANCK_TIME        = 5.391247e-44 // секунда: время прохождения светом планковской длины
	PLANCK_MASS        = 2.176434e-8  // килограмм: масса, при которой комптоновская длина волны равна радиусу Шварцшильда
	PLANCK_ENERGY      = 1.9561e9     // джоуль (эквивалент ~1.22 × 10^19 GeV) — энергия великого объединения
	PLANCK_TEMPERATURE = 1.416784e32  // кельвин: абсолютный горячий предел, температура сразу после Большого взрыва

	// --- Космологические параметры (Planck 2018, ΛCDM модель) ---
	HUBBLE_CONSTANT     = 67.4    // (км/с)/Мпк — постоянная Хаббла, скорость расширения Вселенной сегодня
	DARK_ENERGY_DENSITY = 0.6911  // Ω_Λ — доля тёмной энергии (космологическая постоянная) в критической плотности
	DARK_MATTER_DENSITY = 0.2589  // Ω_c — доля холодной тёмной материи (WIMP, аксионы, стерильные нейтрино)
	BARYON_DENSITY      = 0.0486  // Ω_b — доля барионной (обычной) материи: кварки → протоны, нейтроны
	RADIATION_DENSITY   = 0.000092 // Ω_r — доля излучения: фотоны реликтового фона + безмассовые нейтрино
	CURVATURE_DENSITY   = 0.0007   // Ω_k — отклонение от плоской геометрии (0 = идеально плоская)
	TOTAL_DENSITY       = DARK_ENERGY_DENSITY + DARK_MATTER_DENSITY + BARYON_DENSITY + RADIATION_DENSITY + CURVATURE_DENSITY

	// --- Возраст и масштаб ---
	UNIVERSE_AGE      = 13.787e9       // лет: время, прошедшее с момента Большого взрыва
	OBSERVABLE_RADIUS = 4.4e26         // метров (~46.5 миллиардов световых лет) — сопутствующий радиус
	CMB_TEMPERATURE   = 2.72548        // кельвин: текущая температура реликтового микроволнового фона
	CMB_REDSHIFT      = 1089.0         // z_recombination: красное смещение эпохи рекомбинации

	// --- Константы фундаментальных взаимодействий ---
	ALPHA_EM     = 1.0 / 137.035999084 // постоянная тонкой структуры (электромагнитное взаимодействие)
	ALPHA_STRONG = 0.1183              // константа связи сильного взаимодействия на масштабе массы Z-бозона
	ALPHA_WEAK   = 1.0 / 30.0          // константа связи слабого взаимодействия (~10^-5 на масштабе атомного ядра)
	G_CONSTANT   = 6.67430e-11         // гравитационная постоянная Ньютона (м³/(кг·с²))

	// --- Стандартная модель: массы частиц (МэВ/c²) ---
	// Лептоны
	ELECTRON_MASS      = 0.51099895000 // электрон — стабилен, образует атомные оболочки
	ELECTRON_NEUTRINO_MASS = 1.1e-6   // верхний предел, точное значение неизвестно
	MUON_MASS          = 105.6583745   // мюон — нестабилен, время жизни 2.2 мкс
	MUON_NEUTRINO_MASS = 0.17          // оценка из осцилляций
	TAU_MASS           = 1776.86       // тау-лептон — нестабилен, время жизни 2.9×10^-13 с
	TAU_NEUTRINO_MASS  = 15.5          // оценка из осцилляций

	// Кварки
	UP_QUARK_MASS      = 2.16     // u-кварк: заряд +2/3, масса ~2.16 МэВ
	DOWN_QUARK_MASS    = 4.67     // d-кварк: заряд -1/3, масса ~4.67 МэВ → протон = uud, нейтрон = udd
	STRANGE_QUARK_MASS = 93.4     // s-кварк: заряд -1/3, входит в состав каонов и гиперонов
	CHARM_QUARK_MASS   = 1270.0   // c-кварк: заряд +2/3, открыт в 1974 (J/ψ-мезон)
	BOTTOM_QUARK_MASS  = 4180.0   // b-кварк: заряд -1/3, открыт в 1977 (ϒ-мезон)
	TOP_QUARK_MASS     = 172690.0 // t-кварк: заряд +2/3, самый тяжёлый, открыт в 1995, время жизни ~5×10^-25 с

	// Калибровочные бозоны (переносчики взаимодействий)
	PHOTON_MASS        = 0.0       // фотон — безмассовый, переносчик электромагнитного взаимодействия
	W_BOSON_MASS       = 80379.0   // W± бозоны — переносчики слабого взаимодействия (заряженные токи)
	Z_BOSON_MASS       = 91187.6   // Z⁰ бозон — переносчик слабого взаимодействия (нейтральные токи)
	GLUON_MASS         = 0.0       // глюон — безмассовый, переносчик сильного взаимодействия (8 цветов)

	// Бозон Хиггса (открыт в 2012 на LHC)
	HIGGS_BOSON_MASS   = 125250.0 // МэВ/c² (~125.25 ГэВ/c²) — последняя найденная частица Стандартной модели
	HIGGS_VEV          = 246.22   // ГэВ — вакуумное ожидаемое значение поля Хиггса, определяет массы частиц
	HIGGS_WIDTH        = 4.07e-3  // ГэВ — ширина распада Хиггса (время жизни ~1.6×10^-22 с)

	// --- Астрофизические константы ---
	SOLAR_MASS         = 1.98847e30   // килограмм — масса Солнца
	SOLAR_LUMINOSITY   = 3.828e26     // ватт — светимость Солнца
	SOLAR_RADIUS       = 6.957e8      // метров — радиус Солнца
	SOLAR_TEMPERATURE  = 5778.0       // кельвин — эффективная температура фотосферы Солнца
	EARTH_MASS         = 5.97217e24   // килограмм
	EARTH_RADIUS       = 6.371e6      // метров
	EARTH_TEMPERATURE  = 288.0        // кельвин — средняя температура поверхности Земли
	ASTRONOMICAL_UNIT  = 1.495978707e11 // метров — среднее расстояние Земля-Солнце
	PARSEC             = 3.085677581e16 // метров (~3.26 световых года)

	// --- Универсальные константы ---
	SPEED_OF_LIGHT     = 299792458.0  // м/с (точное значение, определяет метр с 1983 года)
	BOLTZMANN_CONSTANT = 1.380649e-23 // Дж/К (точное значение, определяет кельвин с 2019 года)
	REDUCED_PLANCK     = 1.054571817e-34 // ħ = h/2π (Дж·с)
	PLANCK_CONSTANT    = 6.62607015e-34  // h (Дж·с) — точное значение, определяет килограмм с 2019 года
	ELECTRON_CHARGE    = 1.602176634e-19 // Кл (точное значение, определяет ампер с 2019 года)
	AVOGADRO_NUMBER    = 6.02214076e23   // моль⁻¹ (точное значение, определяет моль с 2019 года)
	FINE_STRUCTURE     = ALPHA_EM        // α ≈ 1/137 — безразмерная, не выводится из теории

	// --- Параметры численной симуляции ---
	GRID_SIZE          = 512     // N³ — количество ячеек в космологической решётке (512³ = 134 миллиона)
	TIME_STEPS         = 1000    // шагов интегрирования от инфляции до формирования первых звёзд
	PARTICLES_PER_CELL = 1000    // типичное количество частиц на ячейку решётки
	MIN_HALO_MASS      = 1e8     // масс Солнца — минимальная масса гало тёмной материи для звездообразования
	GALAXY_COUNT       = 100     // галактик в симулируемом объёме
	STARS_PER_GALAXY   = 1000    // звёзд на галактику (реально ~10¹¹, но мы ограничены памятью)
	PLANETS_PER_STAR   = 5       // планет на звезду (средняя оценка по данным Kepler)
)

// ===================================================================
// КВАНТОВАЯ ТЕОРИЯ ПОЛЯ В ИСКРИВЛЁННОМ ПРОСТРАНСТВЕ-ВРЕМЕНИ
// ===================================================================

// QuantumFieldState описывает квантовое поле в импульсном представлении.
// Каждая мода k представляет собой независимый квантовый гармонический осциллятор.
// В искривлённом пространстве-времени Фридмана моды эволюционируют
// согласно уравнению: φ_k'' + (k² + m²a² - a''/a) φ_k = 0
type QuantumFieldState struct {
	modes        []complex128 // амплитуды мод в импульсном пространстве
	modeCount    int          // полное количество мод до ультрафиолетового обрезания
	cutoffK      float64      // максимальный импульс (планковский масштаб: 2π/l_P)
	effectiveH   float64      // параметр Хаббла H = ȧ/a в данный момент времени
	temperature  float64      // температура поля (для тепловых состояний)
	entropy      float64      // энтропия фон Неймана: S = -Tr(ρ ln ρ)
	particleNumber int        // среднее число частиц, рождённых из вакуума
}

// NewVacuum создаёт чистое вакуумное состояние |0⟩ в пространстве де Ситтера.
// Это основное состояние гамильтониана всех мод.
// Квантовые флуктуации присутствуют даже в вакууме:
// ⟨0|φ_k|0⟩ = 0, но ⟨0|φ_k²|0⟩ = ħ/(2ω_k) ≠ 0
// Эти флуктуации — источник первичных возмущений плотности,
// из которых вырастут все структуры Вселенной.
func NewVacuum(modeCount int) *QuantumFieldState {
	if modeCount <= 0 {
		modeCount = GRID_SIZE * GRID_SIZE * GRID_SIZE
	}
	modes := make([]complex128, modeCount)

	for i := range modes {
		// Частота моды: ω_k = √(k² + m²)
		// Для безмассового поля m=0, ω_k = |k|
		waveNumber := float64(i+1) / float64(modeCount) // нормированный волновой вектор 0..1

		// Амплитуда нулевых колебаний: √(ħ/(2ω_k))
		// В безразмерных единицах: ~1/√(номер_моды)
		amplitude := math.Sqrt(0.5/float64(i+1)) * 1e-15

		// Случайная фаза — квантовая неопределённость
		// В квантовой теории поля вакуум — это суперпозиция всех возможных
		// конфигураций поля, а не "пустота"
		phase := complex(0, rand.Float64()*2*math.Pi)
		modes[i] = amplitude * cmplx.Exp(phase)

		_ = waveNumber // используется в реальной симуляции для ω_k
	}

	return &QuantumFieldState{
		modes:        modes,
		modeCount:    modeCount,
		cutoffK:      2 * math.Pi / PLANCK_LENGTH, // планковское обрезание
		effectiveH:   HUBBLE_CONSTANT * 1000,        // H во время инфляции ~10³⁶ с⁻¹
		temperature:  PLANCK_TEMPERATURE,            // начальная температура = планковская
		entropy:      0,                             // чистое состояние имеет нулевую энтропию
	}
}

// QuantumFluctuation вычисляет вероятность квантового туннелирования
// из ничего ("tunneling from nothing").
// В подходе Виленкина-Хартла-Хокинга: волновая функция Вселенной
// Ψ[h_ij, φ] = ∫ Dg Dφ exp(-S_E[g, φ])
// где S_E — евклидово действие.
// Вероятность рождения Вселенной: P ∝ exp(-|S_E|)
// При S_E ~ -1/(H²G) вероятность становится ~O(1) при H ~ M_Planck.
func (q *QuantumFieldState) QuantumFluctuation() bool {
	// Вычисляем суммарную энергию нулевых колебаний
	totalEnergy := 0.0
	for _, mode := range q.modes {
		totalEnergy += real(mode)*real(mode) + imag(mode)*imag(mode)
	}
	q.entropy = totalEnergy * BOLTZMANN_CONSTANT // S ~ E/T (грубая оценка)

	// Условие начала инфляции: плотность энергии вакуума превышает планковскую
	// ρ_vac > M_Planck⁴ (в естественных единицах)
	criticalDensity := PLANCK_MASS / (PLANCK_LENGTH * PLANCK_LENGTH * PLANCK_LENGTH)
	density := totalEnergy / (PLANCK_LENGTH * PLANCK_LENGTH * PLANCK_LENGTH)

	return density > criticalDensity*0.01 // порог туннелирования
}

// ===================================================================
// ЭПОХА ИНФЛЯЦИИ: ЭКСПОНЕНЦИАЛЬНОЕ РАСШИРЕНИЕ
// ===================================================================

// InflationField описывает инфлатон — скалярное поле, вызвавшее инфляцию.
// Потенциал V(φ) = ½m²φ² (простейшая модель хаотической инфляции Линде)
// или V(φ) = Λ⁴(1 - cos(φ/f)) (аксионная инфляция, естественная инфляция)
type InflationField struct {
	fieldValue    float64 // значение инфлатона φ (в планковских единицах)
	fieldDeriv    float64 // производная dφ/dt
	potentialType int     // 0 = квадратичный, 1 = аксионный, 2 = Starobinsky R²
	slowRollEps   float64 // параметр медленного скатывания ε = (M_P²/2)(V'/V)²
	slowRollEta   float64 // параметр медленного скатывания η = M_P²(V''/V)
	eFolds        float64 // количество e-фолдингов N = ln(a_end/a_start)
	scaleFactor   float64 // масштабный фактор a(t)
}

// NewInflation создаёт начальное состояние инфляции.
// Поле φ ~ несколько M_Planck, медленно скатывается к минимуму потенциала.
// За ~60 e-фолдингов Вселенная расширяется в ~e^60 ≈ 10^26 раз.
func NewInflation() *InflationField {
	return &InflationField{
		fieldValue:    5.0,  // начальное значение φ в планковских единицах (>M_P для хаотической инфляции)
		fieldDeriv:    -0.01, // медленное скатывание
		potentialType: 2,     // модель Старобинского R² (наиболее согласуется с данными Planck)
		slowRollEps:   0.01,  // ε ≪ 1 во время инфляции
		slowRollEta:   0.01,  // η ≪ 1 во время инфляции
		eFolds:        0,     // счётчик e-фолдингов
		scaleFactor:   1e-35, // начальный масштабный фактор ~l_P
	}
}

// Inflate выполняет один шаг инфляционного расширения.
// Уравнение движения: φ'' + 3Hφ' + dV/dφ = 0
// Уравнение Фридмана: H² = (8πG/3)[½φ'² + V(φ)]
// За время dt масштабный фактор увеличивается в exp(H·dt) раз.
func (inf *InflationField) Inflate(dt float64) {
	// Вычисляем потенциал V(φ) и его производные
	var V, dVdphi, d2Vdphi2 float64

	switch inf.potentialType {
	case 0: // Квадратичный потенциал: V = ½m²φ²
		mass := 1e-6 * PLANCK_MASS // масса инфлатона ~10^13 GeV
		V = 0.5 * mass * mass * inf.fieldValue * inf.fieldValue
		dVdphi = mass * mass * inf.fieldValue
		d2Vdphi2 = mass * mass
	case 1: // Аксионный потенциал: V = Λ⁴[1 - cos(φ/f)]
		Lambda := 1e-3 * PLANCK_MASS // масштаб ~10^16 GeV (GUT scale)
		f := 0.1 * PLANCK_MASS       // константа распада аксиона
		V = Lambda * Lambda * Lambda * Lambda * (1 - math.Cos(inf.fieldValue/f))
		dVdphi = Lambda * Lambda * Lambda * Lambda * math.Sin(inf.fieldValue/f) / f
		d2Vdphi2 = Lambda * Lambda * Lambda * Lambda * math.Cos(inf.fieldValue/f) / (f * f)
	case 2: // Модель Старобинского: V ∝ (1 - e^{-√(2/3)φ})²
		norm := 1e-10 * PLANCK_MASS * PLANCK_MASS // нормировка
		expTerm := math.Exp(-math.Sqrt(2.0/3.0) * inf.fieldValue)
		V = norm * (1 - expTerm) * (1 - expTerm)
		dVdphi = norm * 2 * (1 - expTerm) * math.Sqrt(2.0/3.0) * expTerm
		d2Vdphi2 = norm * 2 * (2.0/3.0) * expTerm * (2*expTerm - 1)
	}

	// Параметр Хаббла из уравнения Фридмана (в планковских единицах 8πG=1)
	H := math.Sqrt(V / 3.0)
	if H < 1e-60 {
		H = 1e-60
	}

	// Уравнение движения инфлатона: φ'' + 3Hφ' + dV/dφ = 0
	acceleration := -3*H*inf.fieldDeriv - dVdphi
	inf.fieldDeriv += acceleration * dt
	inf.fieldValue += inf.fieldDeriv * dt

	// Параметры медленного скатывания
	if V > 1e-60 {
		inf.slowRollEps = 0.5 * (dVdphi / V) * (dVdphi / V)
		inf.slowRollEta = d2Vdphi2 / V
	}

	// Расширение: a(t+dt) = a(t) * exp(H·dt)
	eFoldStep := H * dt
	inf.eFolds += eFoldStep
	inf.scaleFactor *= math.Exp(eFoldStep)

	// Инфляция заканчивается, когда ε ~ 1 (нарушение условия медленного скатывания)
	if inf.slowRollEps >= 1.0 && inf.fieldValue > 0 {
		inf.fieldValue *= 0.1 // быстрое скатывание к минимуму
	}
}

// IsInflationOver проверяет условие окончания инфляции.
func (inf *InflationField) IsInflationOver() bool {
	return inf.slowRollEps >= 1.0 && inf.eFolds > 50
}

// ===================================================================
// ЭПОХА РАЗОГРЕВА (REHEATING): РОЖДЕНИЕ ЧАСТИЦ
// ===================================================================

// ParticleType перечисляет все частицы Стандартной модели + тёмная материя
type ParticleType int

const (
	// Кварки (6 типов × 3 цвета = 18 степеней свободы)
	QUARK_UP ParticleType = iota
	QUARK_DOWN
	QUARK_STRANGE
	QUARK_CHARM
	QUARK_BOTTOM
	QUARK_TOP

	// Лептоны (6 типов)
	ELECTRON_PARTICLE
	MUON_PARTICLE
	TAU_PARTICLE
	NEUTRINO_E
	NEUTRINO_MU
	NEUTRINO_TAU

	// Калибровочные бозоны
	PHOTON_PARTICLE
	W_PLUS_BOSON
	W_MINUS_BOSON
	Z_BOSON_PARTICLE
	GLUON_PARTICLE // 8 цветов, но мы упрощаем

	// Хиггс
	HIGGS_PARTICLE

	// За пределами Стандартной модели
	DARK_MATTER_WIMP
	DARK_MATTER_AXION
	STERILE_NEUTRINO
	GRAVITON

	TOTAL_PARTICLE_TYPES
)

// Particle представляет элементарную частицу в фазовом пространстве (x, p).
type Particle struct {
	Type     ParticleType // тип частицы согласно Стандартной модели
	Position [3]float64   // координаты в сопутствующей системе (метры)
	Momentum [3]float64   // импульс (кг·м/с)
	Energy   float64      // полная энергия E = √(p² + m²) (джоули)
	Spin     [3]float64   // вектор спина
	Charge   float64      // электрический заряд (в единицах e)
	Color    int          // цветовой заряд (0-2 для кварков, -1 для лептонов)
	Mass     float64      // масса покоя (кг)
	 Alive   bool         // существует ли частица (не распалась, не аннигилировала)
}

// ParticleUniverse описывает Вселенную, заполненную частицами.
// После reheating Вселенная представляет собой кварк-глюонную плазму
// при температуре ~10^15-10^16 GeV.
type ParticleUniverse struct {
	Particles      []Particle     // все частицы в симуляции
	Temperature    float64        // температура (GeV)
	Time           float64        // время с момента Большого взрыва (секунды)
	ScaleFactor    float64        // масштабный фактор a(t)
	HubbleParam    float64        // параметр Хаббла H(t)
	AntimatterAsym float64        // барионная асимметрия (избыток материи над антиматерией)
	TotalEnergy    float64        // полная энергия всех частиц
	PhotonBaryonRatio float64     // η = n_b/n_γ ~ 6×10^-10
}

// Reheating симулирует рождение частиц после окончания инфляции.
// Механизм: осцилляции инфлатона около минимума потенциала
// параметрически рождают частицы (параметрический резонанс — preheating),
// затем perturbative распад инфлатона завершает reheating.
// Энергия инфлатона переходит в тепловую энергию частиц.
func Reheating(inflation *InflationField) *ParticleUniverse {
	pu := &ParticleUniverse{
		Particles:      make([]Particle, 0, 1000000),
		Temperature:    1e15,                         // GeV — температура великого объединения
		Time:           1e-35,                        // секунд после Большого взрыва
		ScaleFactor:    inflation.scaleFactor,
		HubbleParam:    math.Sqrt(8*math.Pi*G_CONSTANT*PLANCK_ENERGY/3.0) / SPEED_OF_LIGHT,
		AntimatterAsym: 6e-10,                        // наблюдаемая барионная асимметрия
	}

	// Энергия, доступная для рождения частиц: вся энергия инфлатона
	availableEnergy := PLANCK_ENERGY * inflation.scaleFactor * inflation.scaleFactor * inflation.scaleFactor
	numParticles := int(availableEnergy / (PLANCK_MASS * SPEED_OF_LIGHT * SPEED_OF_LIGHT))
	if numParticles > 1000000 {
		numParticles = 1000000
	}
	if numParticles < 1000 {
		numParticles = 1000
	}

	// Рождение частиц: квантовые флуктуации создают пары частица-античастица
	for i := 0; i < numParticles; i++ {
		// Выбираем тип частицы согласно тепловому распределению
		// При высоких температурах все частицы рождаются примерно поровну
		ptype := ParticleType(rand.Intn(int(TOTAL_PARTICLE_TYPES)))

		// Энергия из распределения Бозе-Эйнштейна (для бозонов)
		// или Ферми-Дирака (для фермионов)
		// f(E) = 1/(exp(E/kT) ± 1)
		energy := pu.Temperature * 1e9 * 1.602176634e-19 * rand.ExpFloat64() // джоули
		if energy > PLANCK_ENERGY {
			energy = PLANCK_ENERGY
		}

		// Импульс из релятивистского соотношения
		massKg := particleMassKg(ptype)
		pSquared := (energy/SPEED_OF_LIGHT)*(energy/SPEED_OF_LIGHT) - massKg*massKg*SPEED_OF_LIGHT*SPEED_OF_LIGHT
		if pSquared < 0 {
			pSquared = 0
		}
		momentum := math.Sqrt(pSquared)

		// Случайное направление импульса (изотропия)
		theta := rand.Float64() * math.Pi
		phi := rand.Float64() * 2 * math.Pi
		px := momentum * math.Sin(theta) * math.Cos(phi)
		py := momentum * math.Sin(theta) * math.Sin(phi)
		pz := momentum * math.Cos(theta)

		// Случайное положение в объёме
		volumeRadius := pu.ScaleFactor * OBSERVABLE_RADIUS * 0.01 // малый объём сразу после инфляции
		px_pos := (rand.Float64() - 0.5) * 2 * volumeRadius
		py_pos := (rand.Float64() - 0.5) * 2 * volumeRadius
		pz_pos := (rand.Float64() - 0.5) * 2 * volumeRadius

		p := Particle{
			Type:  ptype,
			Energy: energy,
			Mass:  massKg,
			Momentum: [3]float64{px, py, pz},
			Position: [3]float64{px_pos, py_pos, pz_pos},
			Charge: particleCharge(ptype),
			Color:  particleColor(ptype),
			Alive:  true,
		}
		pu.Particles = append(pu.Particles, p)
	}

	pu.TotalEnergy = availableEnergy
	pu.PhotonBaryonRatio = 6e-10
	return pu
}

// particleMassKg возвращает массу частицы в килограммах.
func particleMassKg(ptype ParticleType) float64 {
	masses := map[ParticleType]float64{
		QUARK_UP:         UP_QUARK_MASS,
		QUARK_DOWN:       DOWN_QUARK_MASS,
		QUARK_STRANGE:    STRANGE_QUARK_MASS,
		QUARK_CHARM:      CHARM_QUARK_MASS,
		QUARK_BOTTOM:     BOTTOM_QUARK_MASS,
		QUARK_TOP:        TOP_QUARK_MASS,
		ELECTRON_PARTICLE: ELECTRON_MASS,
		MUON_PARTICLE:    MUON_MASS,
		TAU_PARTICLE:     TAU_MASS,
		NEUTRINO_E:        ELECTRON_NEUTRINO_MASS,
		NEUTRINO_MU:       MUON_NEUTRINO_MASS,
		NEUTRINO_TAU:      TAU_NEUTRINO_MASS,
		PHOTON_PARTICLE:   PHOTON_MASS,
		W_PLUS_BOSON:      W_BOSON_MASS,
		W_MINUS_BOSON:     W_BOSON_MASS,
		Z_BOSON_PARTICLE:  Z_BOSON_MASS,
		GLUON_PARTICLE:    GLUON_MASS,
		HIGGS_PARTICLE:    HIGGS_BOSON_MASS,
		DARK_MATTER_WIMP:  100000.0, // ~100 GeV
		DARK_MATTER_AXION: 1e-6,     // ~10^-6 eV
		STERILE_NEUTRINO:  10000.0,  // ~10 keV
		GRAVITON:          0,
	}
	// Конвертация МэВ → кг: E = mc², m = E/c²
	// 1 МэВ = 1.78266192e-30 кг
	const MeVToKg = 1.78266192e-30
	if mass, ok := masses[ptype]; ok {
		return mass * MeVToKg
	}
	return 0
}

func particleCharge(ptype ParticleType) float64 {
	charges := map[ParticleType]float64{
		QUARK_UP:         2.0 / 3.0,
		QUARK_CHARM:      2.0 / 3.0,
		QUARK_TOP:        2.0 / 3.0,
		QUARK_DOWN:       -1.0 / 3.0,
		QUARK_STRANGE:    -1.0 / 3.0,
		QUARK_BOTTOM:     -1.0 / 3.0,
		ELECTRON_PARTICLE: -1,
		MUON_PARTICLE:    -1,
		TAU_PARTICLE:     -1,
		W_PLUS_BOSON:     1,
		W_MINUS_BOSON:    -1,
	}
	if charge, ok := charges[ptype]; ok {
		return charge
	}
	return 0
}

func particleColor(ptype ParticleType) int {
	switch ptype {
	case QUARK_UP, QUARK_DOWN, QUARK_STRANGE, QUARK_CHARM, QUARK_BOTTOM, QUARK_TOP:
		return rand.Intn(3) // три цвета: красный, зелёный, синий
	case GLUON_PARTICLE:
		return rand.Intn(8) // восемь цветовых комбинаций
	default:
		return -1 // лептоны и бозоны не имеют цвета
	}
}

// ===================================================================
// ПЕРВИЧНЫЙ НУКЛЕОСИНТЕЗ (Big Bang Nucleosynthesis, BBN)
// ===================================================================

// Nucleosynthesis моделирует образование первых атомных ядер
// через 3-20 минут после Большого взрыва.
// Основные реакции:
// 1. p + n → d + γ (образование дейтерия)
// 2. d + p → ³He + γ
// 3. d + d → ³He + n; d + d → t + p
// 4. ³He + d → ⁴He + p; t + d → ⁴He + n
// 5. ⁴He + t → ⁷Li + γ; ⁴He + ³He → ⁷Be + γ; ⁷Be + n → ⁷Li + p
//
// Результат (по массе): ~75% водорода (¹H), ~25% гелия-4,
// следы дейтерия (~0.002%), ³He (~0.001%), ⁷Li (~10^-9%)
type NucleosynthesisResult struct {
	HydrogenFraction  float64 // доля ¹H
	Helium4Fraction   float64 // доля ⁴He
	DeuteriumFraction float64 // доля ²H (дейтерий)
	Helium3Fraction   float64 // доля ³He
	Lithium7Fraction  float64 // доля ⁷Li
	NeutronFraction   float64 // оставшиеся свободные нейтроны
}

// BigBangNucleosynthesis выполняет расчёт первичного нуклеосинтеза.
// Вход: температура и плотность барионов после аннигиляции e⁺e⁻ пар.
// Выход: распространённость лёгких элементов.
func BigBangNucleosynthesis(temp float64, baryonDensity float64) NucleosynthesisResult {
	// Отношение нейтронов к протонам определяется слабым взаимодействием:
	// n + ν_e ↔ p + e⁻
	// n + e⁺ ↔ p + ν̄_e
	// При закалке (T ~ 0.8 МэВ): n/p = exp(-Δm/T_freeze)
	// где Δm = m_n - m_p = 1.293 МэВ
	deltaM := 1.293 // МэВ
	freezeTemp := 0.8 // МэВ
	neutronProtonRatio := math.Exp(-deltaM / freezeTemp)

	// Дальше нейтроны захватываются в дейтерий, и практически все
	// нейтроны в итоге оказываются в ⁴He
	// Массовая доля ⁴He: Y_p = 2(n/p)/(1 + n/p) на момент нуклеосинтеза
	helium4Mass := 2 * neutronProtonRatio / (1 + neutronProtonRatio)

	// Ограничение на Y_p из наблюдений: 0.245 ± 0.003
	if helium4Mass > 0.26 {
		helium4Mass = 0.25
	}
	if helium4Mass < 0.23 {
		helium4Mass = 0.24
	}

	return NucleosynthesisResult{
		HydrogenFraction:  1.0 - helium4Mass - 0.0001,
		Helium4Fraction:   helium4Mass,
		DeuteriumFraction: 2.5e-5,
		Helium3Fraction:   1.0e-5,
		Lithium7Fraction:  1.0e-10,
		NeutronFraction:   0,
	}
}

// ===================================================================
// РЕКОМБИНАЦИЯ И РЕЛИКТОВОЕ ИЗЛУЧЕНИЕ
// ===================================================================

// Recombination описывает эпоху рекомбинации (z ~ 1089, t ~ 380 000 лет),
// когда электроны связываются с ядрами, образуя нейтральные атомы.
// До рекомбинации: плазма, фотоны рассеиваются на свободных электронах.
// После рекомбинации: Вселенная прозрачна, фотоны свободно распространяются.
// Эти фотоны мы наблюдаем сегодня как реликтовое микроволновое излучение (CMB).
type Recombination struct {
	Redshift        float64 // красное смещение рекомбинации
	Temperature     float64 // температура в момент рекомбинации
	IonizationFrac  float64 // доля ионизированных атомов
	OpticalDepth    float64 // оптическая глубина до поверхности последнего рассеяния
	CMBTemperature  float64 // температура CMB сегодня
}

// PerformRecombination моделирует рекомбинацию водорода.
// Уравнение Саха: n_e²/n_H = (2π m_e kT/h²)^(3/2) exp(-E_ion/kT)
// где E_ion = 13.6 эВ — энергия ионизации водорода.
func PerformRecombination(baryonDensity float64) Recombination {
	// Рекомбинация происходит при T ~ 3000 K (~0.26 эВ)
	recombTemp := 3000.0 // K
	recombRedshift := recombTemp/CMB_TEMPERATURE - 1 // ~1089

	return Recombination{
		Redshift:        recombRedshift,
		Temperature:     recombTemp,
		IonizationFrac:  0.001, // остаточная ионизация ~0.1%
		OpticalDepth:    0.054, // Planck 2018
		CMBTemperature:  CMB_TEMPERATURE,
	}
}

// ===================================================================
// ТЁМНАЯ МАТЕРИЯ И ФОРМИРОВАНИЕ КРУПНОМАСШТАБНОЙ СТРУКТУРЫ
// ===================================================================

// DarkMatterHalo представляет гало тёмной материи — гравитационно связанную
// область с плотностью выше средней по Вселенной.
// Иерархическая кластеризация: малые гало сливаются, образуя крупные.
// Профиль NFW (Navarro-Frenk-White):
// ρ(r) = ρ_s / [(r/r_s)(1 + r/r_s)²]
type DarkMatterHalo struct {
	Mass             float64 // масса гало (в солнечных массах)
	VirialRadius     float64 // радиус вириализации (метры)
	Concentration    float64 // параметр концентрации c = r_vir/r_s
	VelocityDisp     float64 // дисперсия скоростей (м/с)
	FormationRedshift float64 // красное смещение формирования
	Subhalos         []*DarkMatterHalo // субгало (иерархическая структура)
	Position         [3]float64 // координаты центра масс
	AngularMomentum  [3]float64 // угловой момент
	Spin             float64    // безразмерный параметр спина λ
}

// StructureFormation моделирует рост первичных возмущений плотности
// под действием гравитации в расширяющейся Вселенной.
// Начальные условия: гауссово случайное поле с спектром мощности P(k) ∝ k^n_s
// где n_s ≈ 0.965 (Planck 2018) — спектральный индекс.
// Рост возмущений: δ(t) ∝ D(t), где D(t) — фактор роста.
func StructureFormation(dmDensity float64, volumeRadius float64) []*DarkMatterHalo {
	numHalos := int(dmDensity * volumeRadius * volumeRadius * volumeRadius / MIN_HALO_MASS)
	if numHalos > 10000 {
		numHalos = 10000
	}
	if numHalos < 10 {
		numHalos = 10
	}

	halos := make([]*DarkMatterHalo, numHalos)
	for i := 0; i < numHalos; i++ {
		// Масса из функции масс Пресса-Шехтера
		// dN/dM ∝ M^(-1.8) exp(-M/M*)
		mass := math.Pow(rand.Float64(), -1.0) * MIN_HALO_MASS
		if mass > 1e15 {
			mass = 1e15
		}
		if mass < MIN_HALO_MASS {
			mass = MIN_HALO_MASS
		}

		// Соотношение масса-концентрация из N-body симуляций
		concentration := 9.0 * math.Pow(mass/1e12, -0.1)

		// Вириальный радиус из определения: M = (4π/3) r_vir³ ρ_crit Δ_vir
		// ρ_crit = 3H²/(8πG), Δ_vir ≈ 200
		rhoCrit := 3 * (HUBBLE_CONSTANT*1000/PARSEC) * (HUBBLE_CONSTANT*1000/PARSEC) / (8 * math.Pi * G_CONSTANT)
		virialRadius := math.Pow(3*mass*SOLAR_MASS/(4*math.Pi*200*rhoCrit), 1.0/3.0)

		halos[i] = &DarkMatterHalo{
			Mass:              mass,
			VirialRadius:      virialRadius,
			Concentration:     concentration,
			FormationRedshift: rand.Float64() * 10,
			Position: [3]float64{
				(rand.Float64() - 0.5) * 2 * volumeRadius,
				(rand.Float64() - 0.5) * 2 * volumeRadius,
				(rand.Float64() - 0.5) * 2 * volumeRadius,
			},
			Spin: rand.Float64() * 0.1,
		}
	}

	return halos
}

// ===================================================================
// ЗВЁЗДЫ: ЭВОЛЮЦИЯ И НУКЛЕОСИНТЕЗ
// ===================================================================

// Star представляет звезду с её физическими характеристиками
// и эволюционным состоянием.
type Star struct {
	Mass            float64 // масса (в солнечных массах)
	Luminosity      float64 // светимость (в солнечных светимостях)
	Radius          float64 // радиус (в солнечных радиусах)
	Temperature     float64 // эффективная температура фотосферы (K)
	Age             float64 // возраст (миллиарды лет)
	Metallicity     float64 // металличность [Fe/H] — доля элементов тяжелее гелия
	SpectralClass   byte    // O, B, A, F, G, K, M — гарвардская классификация
	LuminosityClass string  // I (сверхгиганты) ... V (главная последовательность)
	EvolutionStage  int     // 0=протозвезда, 1=ГП, 2=субгигант, 3=гигант, 4=сверхгигант, 5=остаток
	Position        [3]float64
	Planets         []*Planet
	HostHalo        *DarkMatterHalo
}

// Planet представляет планету, вращающуюся вокруг звезды.
type Planet struct {
	Mass           float64 // масса (в массах Земли)
	Radius         float64 // радиус (в радиусах Земли)
	OrbitRadius    float64 // большая полуось орбиты (а.е.)
	OrbitPeriod    float64 // орбитальный период (земные годы)
	Eccentricity   float64 // эксцентриситет орбиты
	Temperature    float64 // равновесная температура (K)
	Albedo         float64 // альбедо Бонда (0-1)
	HasWater       bool    // есть ли жидкая вода
	HasAtmosphere  bool    // есть ли атмосфера
	HasMagneticField bool  // есть ли магнитное поле (защита от звёздного ветра)
	Composition    string  // "rocky", "icy", "gas", "iron"
	Type           string  // "terrestrial", "super-earth", "mini-neptune", "gas-giant", "ice-giant"
	HostStar       *Star
}

// CreateStars рождает звёзды в гало тёмной материи согласно
// начальной функции масс (IMF).
func CreateStars(halo *DarkMatterHalo, numStars int) []*Star {
	stars := make([]*Star, numStars)

	for i := 0; i < numStars; i++ {
		// Начальная функция масс Круппы (Kroupa IMF):
		// dN/dM ∝ M^(-0.3) для 0.01-0.08 M☉
		// dN/dM ∝ M^(-1.3) для 0.08-0.5 M☉
		// dN/dM ∝ M^(-2.3) для 0.5-150 M☉
		randVal := rand.Float64()
		var mass float64
		if randVal < 0.1 {
			mass = 0.01 + rand.Float64()*0.07 // коричневые карлики
		} else if randVal < 0.5 {
			mass = 0.08 + rand.Float64()*0.42 // красные карлики
		} else {
			mass = 0.5 + math.Pow(rand.Float64(), -0.8) // массивные звёзды
			if mass > 150 {
				mass = 150
			}
		}

		// Соотношение масса-светимость для главной последовательности
		var luminosity float64
		if mass < 0.43 {
			luminosity = 0.23 * math.Pow(mass, 2.3)
		} else if mass < 2.0 {
			luminosity = math.Pow(mass, 4.0)
		} else if mass < 20 {
			luminosity = 1.4 * math.Pow(mass, 3.5)
		} else {
			luminosity = 32000 * mass
		}

		// Соотношение масса-радиус
		var radius float64
		if mass < 1.0 {
			radius = math.Pow(mass, 0.8)
		} else {
			radius = math.Pow(mass, 0.57)
		}

		// Температура из закона Стефана-Больцмана: L = 4πR²σT⁴
		stefanBoltzmann := 5.670374419e-8
		temperature := math.Pow(
			luminosity*SOLAR_LUMINOSITY/(4*math.Pi*radius*radius*SOLAR_RADIUS*SOLAR_RADIUS*stefanBoltzmann),
			0.25,
		)

		// Спектральный класс по гарвардской системе
		spectralClass := byte('M')
		if temperature > 30000 {
			spectralClass = 'O'
		} else if temperature > 10000 {
			spectralClass = 'B'
		} else if temperature > 7500 {
			spectralClass = 'A'
		} else if temperature > 6000 {
			spectralClass = 'F'
		} else if temperature > 5200 {
			spectralClass = 'G'
		} else if temperature > 3700 {
			spectralClass = 'K'
		}

		// Возраст: чем массивнее звезда, тем короче жизнь
		// Время жизни на ГП: t ∝ M/L ∝ M^(-2.5)
		mainSequenceLifetime := 10.0 * math.Pow(mass, -2.5) // миллиарды лет
		if mainSequenceLifetime > 13.8 {
			mainSequenceLifetime = 13.8
		}
		age := rand.Float64() * mainSequenceLifetime

		// Эволюционная стадия
		evolutionStage := 1 // главная последовательность
		if age > mainSequenceLifetime*0.9 {
			evolutionStage = 2 // субгигант
		}
		if age > mainSequenceLifetime {
			evolutionStage = 3 // красный гигант
		}
		if mass > 8 && age > mainSequenceLifetime {
			evolutionStage = 4 // сверхгигант
		}

		star := &Star{
			Mass:            mass,
			Luminosity:      luminosity,
			Radius:          radius,
			Temperature:     temperature,
			Age:             age,
			Metallicity:     rand.NormFloat64()*0.5 - 0.3, // распределение металличности
			SpectralClass:   spectralClass,
			LuminosityClass: "V",
			EvolutionStage:  evolutionStage,
			Position: [3]float64{
				halo.Position[0] + (rand.Float64()-0.5)*halo.VirialRadius,
				halo.Position[1] + (rand.Float64()-0.5)*halo.VirialRadius,
				halo.Position[2] + (rand.Float64()-0.5)*halo.VirialRadius,
			},
			Planets:  make([]*Planet, 0),
			HostHalo: halo,
		}
		stars[i] = star
	}

	return stars
}

// CreatePlanets создаёт планетную систему вокруг звезды.
func CreatePlanets(star *Star) []*Planet {
	// Количество планет из распределения Kepler
	numPlanets := 1 + rand.Intn(8) // 1-8 планет
	if star.Mass > 2.0 {
		numPlanets = rand.Intn(5) // массивные звёзды имеют меньше планет
	}

	planets := make([]*Planet, numPlanets)

	for i := 0; i < numPlanets; i++ {
		// Орбитальный радиус: закон Тициуса-Боде (модифицированный)
		orbitRadius := 0.05 + float64(i)*0.3 + rand.Float64()*0.2
		if i > 3 {
			orbitRadius = 1.5 + float64(i-3)*2.0 + rand.Float64()*3.0
		}

		// Равновесная температура
		// T_eq = T_star × √(R_star/(2a)) × (1-A)^(1/4)
		albedo := 0.3 + rand.Float64()*0.4
		eqTemp := star.Temperature * math.Sqrt(star.Radius*SOLAR_RADIUS/(2*orbitRadius*ASTRONOMICAL_UNIT))
		eqTemp *= math.Pow(1-albedo, 0.25)
		if eqTemp < 10 {
			eqTemp = 10
		}
		if eqTemp > 2000 {
			eqTemp = 2000
		}

		// Масса планеты (распределение из наблюдений)
		planetType := rand.Float64()
		var mass, radius float64
		var composition, ptype string

		if planetType < 0.3 {
			// Землеподобные (скалистые)
			mass = 0.1 + rand.Float64()*2.0
			radius = math.Pow(mass, 0.3)
			composition = "rocky"
			ptype = "terrestrial"
		} else if planetType < 0.6 {
			// Суперземли
			mass = 2.0 + rand.Float64()*8.0
			radius = math.Pow(mass, 0.25)
			composition = "rocky"
			ptype = "super-earth"
		} else if planetType < 0.85 {
			// Мини-нептуны
			mass = 5.0 + rand.Float64()*30.0
			radius = 2.0 + rand.Float64()*2.0
			composition = "icy"
			ptype = "mini-neptune"
		} else {
			// Газовые гиганты
			mass = 30.0 + rand.Float64()*300.0
			radius = 8.0 + rand.Float64()*10.0
			composition = "gas"
			ptype = "gas-giant"
		}

		// Орбитальный период: T² ∝ a³ (третий закон Кеплера)
		orbitPeriod := math.Sqrt(orbitRadius*orbitRadius*orbitRadius/star.Mass) // годы

		// Зона обитаемости: 200 K < T_eq < 350 K для жидкой воды
		hasWater := false
		hasAtmosphere := false
		hasMagneticField := false

		if eqTemp > 200 && eqTemp < 350 && ptype == "terrestrial" || ptype == "super-earth" {
			hasWater = rand.Float64() > 0.4
			hasAtmosphere = rand.Float64() > 0.3
			hasMagneticField = rand.Float64() > 0.5
		}

		planet := &Planet{
			Mass:            mass,
			Radius:          radius,
			OrbitRadius:     orbitRadius,
			OrbitPeriod:     orbitPeriod,
			Eccentricity:    rand.Float64() * 0.1,
			Temperature:     eqTemp,
			Albedo:          albedo,
			HasWater:        hasWater,
			HasAtmosphere:   hasAtmosphere,
			HasMagneticField: hasMagneticField,
			Composition:     composition,
			Type:            ptype,
			HostStar:        star,
		}
		planets[i] = planet
	}

	return planets
}

// ===================================================================
// АБИОГЕНЕЗ: ВОЗНИКНОВЕНИЕ ЖИЗНИ ИЗ НЕЖИВОЙ МАТЕРИИ
// ===================================================================

// OrganicMolecule представляет органическую молекулу — строительный блок жизни.
type OrganicMolecule struct {
	Name        string
	Formula     string
	Complexity  int     // количество атомов
	Energy      float64 // свободная энергия Гиббса образования (кДж/моль)
	Concentration float64 // концентрация в первичном бульоне
}

// Protocell — протоклетка: самовоспроизводящаяся везикула
// с примитивным метаболизмом и генетическим материалом (РНК).
type Protocell struct {
	Membrane      bool              // есть ли липидная мембрана
	RNA           []byte            // последовательность рибонуклеотидов
	RNALength     int               // длина РНК
	Metabolites   map[string]float64 // концентрации метаболитов
	GrowthRate    float64            // скорость роста
	DivisionProb  float64            // вероятность деления в единицу времени
	Alive         bool
}

// Abiogenesis моделирует возникновение жизни из неорганических соединений.
//
// Ключевые этапы:
// 1. Синтез органических мономеров (аминокислоты, нуклеотиды) из CO₂, H₂O, NH₃, CH₄, H₂S
//    - Эксперимент Миллера-Юри (1953): разряды в восстановительной атмосфере → аминокислоты
//    - Гидротермальные источники: H₂ + CO₂ → органические кислоты
//    - Космическая пыль: доставка органики на раннюю Землю
//
// 2. Полимеризация: мономеры → полимеры (белки, РНК)
//    - Высушивание на глинистых минералах (монтмориллонит)
//    - Катализ ионами металлов (Fe²⁺, Mg²⁺, Zn²⁺)
//
// 3. Возникновение самовоспроизводящихся систем
//    - РНК-мир: рибозимы катализируют собственную репликацию
//    - Липосомы: спонтанное формирование везикул из жирных кислот
//    - Протоклетки: инкапсуляция РНК в липосомы
func Abiogenesis(planet *Planet) bool {
	// Необходимые условия для абиогенеза
	if !planet.HasWater {
		return false // нет жидкой воды — нет жизни (как мы её знаем)
	}
	if !planet.HasAtmosphere {
		return false // нет атмосферы — нет защиты от УФ и космических лучей
	}
	if planet.Temperature < 200 || planet.Temperature > 400 {
		return false // вне диапазона стабильности сложных органических молекул
	}

	// Оценка количества попыток самосборки протоклетки
	// На ранней Земле: ~10^40 молекул в океанах, ~10^8 лет,
	// ~10^13 попыток в секунду → ~10^60 попыток всего
	oceanMass := planet.Mass * EARTH_MASS * 0.001 // ~0.1% массы планеты — вода
	moleculesInOcean := oceanMass / 0.018 * AVOGADRO_NUMBER // масса воды / молярная масса × число Авогадро
	attemptsPerMolecule := 1e13 // попыток в секунду на молекулу
	yearsAvailable := 5e8       // ~500 миллионов лет до появления жизни на Земле
	secondsPerYear := 3.15576e7
	totalAttempts := moleculesInOcean * attemptsPerMolecule * yearsAvailable * secondsPerYear

	// Вероятность успешной самосборки протоклетки за одну попытку
	// Оценка снизу: ~10^-40 (очень грубая)
	// Некоторые оценки дают 10^-30 для простейшей РНК-полимеразы
	probPerAttempt := 1e-35

	// Полная вероятность
	successProb := 1.0 - math.Exp(-totalAttempts*probPerAttempt)

	// Жизнь возникает, если вероятность > 0.5
	if successProb > 0.5 {
		// Создаём первую протоклетку
		firstCell := Protocell{
			Membrane:     true,
			RNA:          make([]byte, 40+rand.Intn(60)),
			RNALength:    40 + rand.Intn(60),
			Metabolites:  make(map[string]float64),
			GrowthRate:   0.001,
			DivisionProb: 0.01,
			Alive:        true,
		}
		// Случайная последовательность РНК (A, U, G, C)
		for j := range firstCell.RNA {
			firstCell.RNA[j] = byte(rand.Intn(4))
		}
		_ = firstCell
		return true
	}

	return false
}

// ===================================================================
// БИОЛОГИЧЕСКАЯ ЭВОЛЮЦИЯ
// ===================================================================

// Organism представляет биологический организм с генетическим кодом.
type Organism struct {
	Species      string  // название вида
	GenomeSize   int     // размер генома (миллионы пар оснований)
	Complexity   float64 // морфологическая сложность (0=LUCA, 1=человек)
	Intelligence float64 // когнитивные способности (0=нет, 1=человек)
	Population   int64   // размер популяции
	Generation   int64   // номер поколения
	BrainSize    float64 // объём мозга (см³)
	BodyMass     float64 // масса тела (кг)
	IsExtinct    bool
}

// LUCA — Last Universal Common Ancestor, последний универсальный общий предок.
// Жил ~3.5-3.8 миллиарда лет назад. Уже имел ДНК, РНК, белки,
// рибосомы, АТФ-синтазу, генетический код.
func CreateLUCA() *Organism {
	return &Organism{
		Species:      "LUCA",
		GenomeSize:   2,    // ~2 миллиона пар оснований (минимальный геном)
		Complexity:   0.001,
		Intelligence: 0,
		Population:   1000,
		Generation:   0,
		BrainSize:    0,
		BodyMass:     1e-15, // ~пикограмм (одноклеточное)
	}
}

// Evolve выполняет один шаг эволюции популяции.
// Включает: мутации, генетический дрейф, естественный отбор, видообразование.
func Evolve(org *Organism, environment Planet) *Organism {
	if org.IsExtinct {
		return org
	}

	// Мутации: случайные изменения генома
	// Скорость мутации: ~10^-8 на нуклеотид на поколение
	mutationRate := 1e-8 * float64(org.GenomeSize*1e6)
	mutations := 0
	for i := 0.0; i < mutationRate; i++ {
		if rand.Float64() < mutationRate/float64(org.GenomeSize*1e6) {
			mutations++
		}
	}

	// Полезные мутации редки (~1%)
	beneficialMutations := 0
	for i := 0; i < mutations; i++ {
		if rand.Float64() < 0.01 {
			beneficialMutations++
		}
	}

	// Рост сложности от полезных мутаций
	complexityIncrease := float64(beneficialMutations) * 0.0001
	org.Complexity += complexityIncrease

	// Горизонтальный перенос генов (для прокариот)
	if org.Complexity < 0.1 {
		org.GenomeSize += rand.Intn(10)
	}

	// Естественный отбор: выживают наиболее приспособленные
	// Приспособленность = функция от сложности и соответствия среде
	fitness := org.Complexity * 100
	if environment.Temperature < 250 || environment.Temperature > 350 {
		fitness *= 0.5 // экстремальная температура снижает приспособленность
	}
	if !environment.HasWater {
		fitness *= 0.1
	}

	// Популяционная динамика
	carryingCapacity := int64(1e9 * fitness)
	if carryingCapacity < 100 {
		carryingCapacity = 100
	}
	if carryingCapacity > 1e12 {
		carryingCapacity = 1e12
	}

	if org.Population > carryingCapacity {
		org.Population = carryingCapacity
	} else if org.Population < carryingCapacity {
		org.Population += int64(float64(org.Population) * 0.1) // рост 10%
	}

	// Генетический дрейф в малых популяциях
	if org.Population < 10000 {
		org.Complexity += rand.NormFloat64() * 0.001
	}

	// Развитие интеллекта
	// Появляется при достаточной сложности (многоклеточные)
	if org.Complexity > 0.3 {
		org.Intelligence += 0.00001 * float64(beneficialMutations)
	}

	// Развитие мозга
	if org.Intelligence > 0.01 {
		org.BrainSize = math.Pow(org.Intelligence, 2) * 1350 // масштабирование к человеческому мозгу
	}

	// Рост тела
	org.BodyMass = 1e-15 + org.Complexity*org.Complexity*70 // масштабирование к 70 кг

	// Видообразование: новые виды возникают при накоплении различий
	org.Generation++
	if org.Generation%100000 == 0 {
		org.Species = evolveSpeciesName(org)
	}

	// Вымирание: если приспособленность слишком низкая
	if fitness < 0.01 && rand.Float64() < 0.1 {
		org.IsExtinct = true
	}

	// Кембрийский взрыв (~540 млн лет назад): резкий рост разнообразия
	if org.Complexity > 0.2 && org.Complexity < 0.3 && rand.Float64() < 0.3 {
		org.Complexity += 0.05
		org.GenomeSize *= 2
	}

	// Массовые вымирания (каждые ~100 млн лет)
	if org.Generation%100000000 == 0 {
		if rand.Float64() < 0.7 { // 70% видов вымирает
			org.Population /= 10
			if org.Population < 100 {
				org.IsExtinct = true
			}
		}
	}

	return org
}

func evolveSpeciesName(org *Organism) string {
	names := []string{
		"LUCA", "Prokaryota", "Eukaryota", "Metazoa", "Bilateria",
		"Chordata", "Vertebrata", "Tetrapoda", "Amniota", "Mammalia",
		"Primates", "Hominidae", "Homo Habilis", "Homo Erectus",
		"Homo Neanderthalensis", "Homo Sapiens",
	}

	idx := int(org.Complexity * float64(len(names)-1))
	if idx >= len(names) {
		idx = len(names) - 1
	}
	if idx < 0 {
		idx = 0
	}
	return names[idx]
}

// ===================================================================
// ТЕХНОЛОГИЧЕСКАЯ ЦИВИЛИЗАЦИЯ
// ===================================================================

// Technology представляет технологическое изобретение.
type Technology struct {
	Name        string
	Year        int64   // год изобретения от начала цивилизации
	Complexity  float64 // сложность
	PrereqTech  []string // необходимые предыдущие технологии
	Impact      float64  // влияние на прогресс
}

// Civilization описывает технологическую цивилизацию.
type Civilization struct {
	Species          *Organism
	Planet           *Planet
	TechLevel        float64        // 0=каменный век, 1=космическая эра
	Population       int64
	Year             int64          // годы от основания
	Technologies     []Technology
	HasWriting       bool
	HasScience       bool
	HasComputers     bool
	HasInternet      bool
	HasAI            bool
	HasGo            bool
	GoProgrammers    []string
	KardashevScale   float64       // шкала Кардашёва (0=планетарная, 1=звёздная)
	SurvivedFilter   bool          // прошла ли Великий фильтр
}

// DevelopCivilization симулирует развитие цивилизации от аграрной до космической.
func DevelopCivilization(org *Organism, planet *Planet) *Civilization {
	civ := &Civilization{
		Species:        org,
		Planet:         planet,
		TechLevel:      0.01,
		Population:     org.Population / 100, // начальная популяция разумных существ
		Year:           0,
		Technologies:   make([]Technology, 0),
		GoProgrammers:  make([]string, 0),
		KardashevScale: 0.0,
		SurvivedFilter: true,
	}

	// Технологическое древо
	techTree := []Technology{
		{Name: "Каменные орудия", Year: 0, Complexity: 0.001, Impact: 0.1},
		{Name: "Огонь", Year: 1000, Complexity: 0.002, Impact: 0.15},
		{Name: "Язык", Year: 5000, Complexity: 0.005, Impact: 0.2},
		{Name: "Земледелие", Year: 10000, Complexity: 0.01, Impact: 0.3},
		{Name: "Письменность", Year: 15000, Complexity: 0.02, Impact: 0.4},
		{Name: "Математика", Year: 16000, Complexity: 0.03, Impact: 0.3},
		{Name: "Металлургия", Year: 17000, Complexity: 0.04, Impact: 0.25},
		{Name: "Печатный станок", Year: 18500, Complexity: 0.06, Impact: 0.5},
		{Name: "Паровой двигатель", Year: 19000, Complexity: 0.08, Impact: 0.4},
		{Name: "Электричество", Year: 19200, Complexity: 0.1, Impact: 0.5},
		{Name: "Телефон", Year: 19300, Complexity: 0.12, Impact: 0.3},
		{Name: "Радио", Year: 19350, Complexity: 0.13, Impact: 0.3},
		{Name: "Телевидение", Year: 19400, Complexity: 0.14, Impact: 0.25},
		{Name: "Ядерная энергия", Year: 19450, Complexity: 0.2, Impact: 0.6},
		{Name: "Транзистор", Year: 19470, Complexity: 0.25, Impact: 0.7},
		{Name: "Компьютер", Year: 19480, Complexity: 0.3, Impact: 0.8},
		{Name: "Интернет", Year: 19500, Complexity: 0.35, Impact: 0.9},
		{Name: "Смартфон", Year: 19520, Complexity: 0.4, Impact: 0.6},
		{Name: "Искусственный интеллект", Year: 19530, Complexity: 0.5, Impact: 1.0},
		{Name: "Go (язык программирования)", Year: 19540, Complexity: 0.45, Impact: 0.5},
		{Name: "Квантовый компьютер", Year: 19550, Complexity: 0.6, Impact: 0.8},
		{Name: "Термоядерный синтез", Year: 19560, Complexity: 0.7, Impact: 0.9},
		{Name: "Межзвёздные перелёты", Year: 19600, Complexity: 0.9, Impact: 1.5},
	}

	// Симуляция технологического прогресса
	for _, tech := range techTree {
		civ.Year = tech.Year
		civ.TechLevel = math.Max(civ.TechLevel, tech.Complexity)

		// Рост населения с технологическим прогрессом
		civ.Population = int64(float64(civ.Population) * (1 + tech.Impact*0.1))
		if civ.Population > 1e10 {
			civ.Population = 1e10
		}

		// Принятие технологий
		if tech.Name == "Письменность" {
			civ.HasWriting = true
		}
		if tech.Name == "Математика" {
			civ.HasScience = true
		}
		if tech.Name == "Компьютер" {
			civ.HasComputers = true
		}
		if tech.Name == "Интернет" {
			civ.HasInternet = true
		}
		if tech.Name == "Искусственный интеллект" {
			civ.HasAI = true
		}
		if tech.Name == "Go (язык программирования)" {
			civ.HasGo = true
			// Создатели Go
			civ.GoProgrammers = append(civ.GoProgrammers, "Robert Griesemer", "Rob Pike", "Ken Thompson")
			// Сообщество
			for i := 0; i < 100; i++ {
				civ.GoProgrammers = append(civ.GoProgrammers, fmt.Sprintf("Gopher_%d", i))
			}
		}

		civ.Technologies = append(civ.Technologies, tech)
	}

	// Шкала Кардашёва
	totalEnergy := float64(civ.Population) * 10000 // ~10 кВт на человека
	solarEnergy := SOLAR_LUMINOSITY
	civ.KardashevScale = math.Log10(totalEnergy/solarEnergy)/10 + 0.7

	// Великий фильтр: ядерная война, климатическая катастрофа, и т.д.
	if civ.TechLevel > 0.4 && civ.TechLevel < 0.8 {
		filterProbability := 0.3 // 30% шанс не пройти фильтр
		if rand.Float64() < filterProbability {
			civ.SurvivedFilter = false
			civ.Population /= 1000
			civ.TechLevel *= 0.1
		}
	}

	return civ
}

// ===================================================================
// ЯЗЫК GO: "HELLO WORLD"
// ===================================================================

// GoProgram представляет программу на языке Go, написанную
// представителем симулированной цивилизации.
type GoProgram struct {
	Source   string
	Output   string
	Author   string
	Compiled bool
	Executed bool
}

// WriteHelloWorld — финальный акт творения.
// Один из программистов пишет "hello world".
// Этот момент — цель всей симуляции.
// После 13.8 миллиардов лет космической эволюции,
// после формирования галактик, звёзд, планет,
// после возникновения жизни и миллиарда лет эволюции,
// после тысяч лет технологического прогресса —
// одна строка кода воплощает всю историю Вселенной.
func WriteHelloWorld(civ *Civilization) *GoProgram {
	if !civ.HasGo || len(civ.GoProgrammers) == 0 {
		return nil
	}

	// Выбираем случайного программиста
	authorIdx := rand.Intn(len(civ.GoProgrammers))
	author := civ.GoProgrammers[authorIdx]

	program := &GoProgram{
		Source: `package main

import "fmt"

func main() {
    fmt.Println("hello world")
}`,
		Author: author,
	}

	// Компиляция: go build
	program.Compiled = true

	// Исполнение
	program.Output = "hello world"
	program.Executed = true

	return program
}

// ===================================================================
// ВИЗУАЛИЗАЦИЯ: ПРЯМОЙ ВЫВОД В ВИДЕОПАМЯТЬ (БЕЗ ОС!)
// ===================================================================

var (
	vgaReady int32
	vgaMutex sync.Mutex
)

// vgaPrintHelloWorld выводит результат симуляции непосредственно
// в текстовый видеобуфер VGA по адресу 0xB8000.
// Это работает только на bare metal (без операционной системы)
// или в эмуляторе с прямым доступом к памяти.
func vgaPrintHelloWorld() {
	const VGA_ADDRESS = uintptr(0xB8000)
	const VGA_WIDTH = 80
	const VGA_HEIGHT = 25

	msg := "hello world"

	if !atomic.CompareAndSwapInt32(&vgaReady, 0, 1) {
		return
	}
	vgaMutex.Lock()
	defer vgaMutex.Unlock()

	// Очистка экрана
	for row := 0; row < VGA_HEIGHT; row++ {
		for col := 0; col < VGA_WIDTH; col++ {
			offset := (row*VGA_WIDTH + col) * 2
			ptr := (*[2]byte)(unsafe.Pointer(VGA_ADDRESS + uintptr(offset)))
			ptr[0] = ' '
			ptr[1] = 0x07 // светло-серый на чёрном
		}
	}

	// Вывод сообщения в центре экрана
	row := VGA_HEIGHT / 2
	col := (VGA_WIDTH - len(msg)) / 2

	for i, c := range []byte(msg) {
		offset := (row*VGA_WIDTH + col + i) * 2
		ptr := (*[2]byte)(unsafe.Pointer(VGA_ADDRESS + uintptr(offset)))
		ptr[0] = c
		ptr[1] = 0x0F // ярко-белый на чёрном
	}

	runtime.KeepAlive(&msg)
}

// rdtsc читает счётчик тактов процессора (Time Stamp Counter).
// Используется для измерения времени и как источник аппаратной энтропии.
//
//go:noescape
func rdtsc() uint64

// ===================================================================
// ГЛАВНЫЙ ЦИКЛ СИМУЛЯЦИИ
// ===================================================================

func init() {
	// Инициализация генератора случайных чисел
	// В реальной симуляции здесь должен быть квантовый генератор
	rand.Seed(time.Now().UnixNano())
}

func main() {
	// Заголовок симуляции
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   ПОЛНАЯ ФИЗИЧЕСКАЯ СИМУЛЯЦИЯ ВСЕЛЕННОЙ                    ║")
	fmt.Println("║   От Большого взрыва до Hello World                        ║")
	fmt.Println("║   Квантовая теория поля + ОТО + Стандартная модель         ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Запуск симуляции с фундаментальными константами:")
	fmt.Printf("  Постоянная тонкой структуры: α = %.9f\n", ALPHA_EM)
	fmt.Printf("  Постоянная Хаббла: H₀ = %.1f км/с/Мпк\n", HUBBLE_CONSTANT)
	fmt.Printf("  Планковская длина: l_P = %.6e м\n", PLANCK_LENGTH)
	fmt.Printf("  Планковское время: t_P = %.6e с\n", PLANCK_TIME)
	fmt.Printf("  Планковская температура: T_P = %.6e K\n", PLANCK_TEMPERATURE)
	fmt.Println()

	// ========================================
	// УРОВЕНЬ 0: КВАНТОВЫЙ ВАКУУМ
	// ========================================
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("[Эпоха 0] Квантовый вакуум — рождение из ничего")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  Теория: Квантовая гравитация (петлевая / струнная)")
	fmt.Println("  Состояние: |Ψ⟩ — волновая функция Вселенной")
	fmt.Println("  Описание: Чистое вакуумное состояние в гильбертовом пространстве")
	fmt.Println()

	vacuum := NewVacuum(GRID_SIZE * GRID_SIZE * GRID_SIZE)
	fmt.Printf("  Мод в импульсном пространстве: %d\n", vacuum.modeCount)
	fmt.Printf("  Ультрафиолетовое обрезание: k_max = 2π/l_P ≈ %.3e м⁻¹\n", vacuum.cutoffK)
	fmt.Printf("  Начальная энтропия: S = %.6f (чистое состояние)\n", vacuum.entropy)
	fmt.Println()

	// Ожидание квантовой флуктуации
	fmt.Println("  Ожидание квантовой флуктуации...")
	fluctuations := 0
	inflated := false
	for !inflated {
		inflated = vacuum.QuantumFluctuation()
		fluctuations++
		if fluctuations%100000 == 0 {
			fmt.Printf("  ... попыток туннелирования: %d (планковское время t = %d t_P)\n",
				fluctuations, fluctuations)
		}
	}
	fmt.Printf("  ✦ ТУННЕЛИРОВАНИЕ ИЗ НИЧЕГО ПРОИЗОШЛО ✦ (попытка %d)\n", fluctuations)
	fmt.Printf("  Плотность энергии вакуума превысила критическую\n")
	fmt.Println()

	// ========================================
	// УРОВЕНЬ 1: ИНФЛЯЦИЯ
	// ========================================
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("[Эпоха 1] Инфляция — экспоненциальное расширение")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  T = 10^-36 … 10^-32 с после Большого взрыва")
	fmt.Println("  Модель: R²-инфляция Старобинского (согласуется с Planck 2018)")
	fmt.Println("  Описание: Инфлатонное поле φ медленно скатывается к минимуму")
	fmt.Println("            потенциала, вызывая экспоненциальное расширение")
	fmt.Println()

	inflation := NewInflation()
	fmt.Printf("  Начальное значение инфлатона: φ₀ = %.2f M_P\n", inflation.fieldValue)
	fmt.Printf("  Начальный масштабный фактор: a₀ = %.3e м\n", inflation.scaleFactor)

	// Симуляция инфляции
	dt := 1e-38 // секунды — шаг интегрирования
	steps := 0
	for !inflation.IsInflationOver() && steps < 100000 {
		inflation.Inflate(dt)
		steps++
		if steps%10000 == 0 {
			fmt.Printf("  ... e-фолдингов: %.1f, масштабный фактор: a = %.3e м\n",
				inflation.eFolds, inflation.scaleFactor)
		}
	}
	fmt.Printf("  Инфляция завершена после %d шагов\n", steps)
	fmt.Printf("  Всего e-фолдингов: N = %.1f\n", inflation.eFolds)
	fmt.Printf("  Масштабный фактор после инфляции: a = %.3e м\n", inflation.scaleFactor)
	fmt.Printf("  Расширение в e^N ≈ %.1e раз\n", math.Exp(inflation.eFolds))
	fmt.Printf("  Параметр ε на выходе: %.4f\n", inflation.slowRollEps)
	fmt.Println()

	// ========================================
	// УРОВЕНЬ 2: REHEATING — РОЖДЕНИЕ ЧАСТИЦ
	// ========================================
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("[Эпоха 2] Reheating — рождение частиц")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  T = 10^-32 … 10^-12 с после Большого взрыва")
	fmt.Println("  Описание: Энергия инфлатона переходит в частицы Стандартной")
	fmt.Println("            модели через параметрический резонанс и распад")
	fmt.Println()

	particleUniverse := Reheating(inflation)
	fmt.Printf("  Частиц рождено: %d\n", len(particleUniverse.Particles))
	fmt.Printf("  Температура после reheating: T ≈ %.1e ГэВ\n", particleUniverse.Temperature)
	fmt.Printf("  Масштабный фактор: a = %.3e м\n", particleUniverse.ScaleFactor)
	fmt.Printf("  Барионная асимметрия: η = %.1e\n", particleUniverse.AntimatterAsym)

	// Подсчёт частиц по типам
	particleCounts := make(map[ParticleType]int)
	for _, p := range particleUniverse.Particles {
		particleCounts[p.Type]++
	}
	fmt.Println("  Распределение частиц по типам (первые 10):")
	count := 0
	for ptype, cnt := range particleCounts {
		if count < 10 {
			fmt.Printf("    %d: %d шт.\n", ptype, cnt)
			count++
		}
	}
	fmt.Println()

	// ========================================
	// УРОВЕНЬ 3: НУКЛЕОСИНТЕЗ
	// ========================================
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("[Эпоха 3] Первичный нуклеосинтез (BBN)")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  T = 3 … 20 минут после Большого взрыва")
	fmt.Println("  T_физическая = 10^9 … 10^7 K")
	fmt.Println()

	bbnResult := BigBangNucleosynthesis(1e9, BARYON_DENSITY)
	fmt.Printf("  Результаты нуклеосинтеза:\n")
	fmt.Printf("    Водород (¹H):    %.1f%%\n", bbnResult.HydrogenFraction*100)
	fmt.Printf("    Гелий-4 (⁴He):   %.1f%%\n", bbnResult.Helium4Fraction*100)
	fmt.Printf("    Дейтерий (²H):   %.4f%%\n", bbnResult.DeuteriumFraction*100)
	fmt.Printf("    Гелий-3 (³He):   %.4f%%\n", bbnResult.Helium3Fraction*100)
	fmt.Printf("    Литий-7 (⁷Li):   %.10f%%\n", bbnResult.Lithium7Fraction*100)
	fmt.Println("  (Согласуется с наблюдениями в пределах погрешности)")
	fmt.Println()

	// ========================================
	// УРОВЕНЬ 4: РЕКОМБИНАЦИЯ И CMB
	// ========================================
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("[Эпоха 4] Рекомбинация — рождение реликтового излучения")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  T = 380 000 лет после Большого взрыва")
	fmt.Println()

	recomb := PerformRecombination(BARYON_DENSITY)
	fmt.Printf("  Красное смещение рекомбинации: z = %.0f\n", recomb.Redshift)
	fmt.Printf("  Температура рекомбинации: T = %.0f K\n", recomb.Temperature)
	fmt.Printf("  Остаточная ионизация: %.2f%%\n", recomb.IonizationFrac*100)
	fmt.Printf("  Оптическая глубина: τ = %.4f\n", recomb.OpticalDepth)
	fmt.Printf("  CMB сегодня: T₀ = %.5f K\n", recomb.CMBTemperature)
	fmt.Println("  Вселенная стала прозрачной — фотоны свободно распространяются")
	fmt.Println()

	// ========================================
	// УРОВЕНЬ 5: ТЁМНЫЕ ВЕКА И ФОРМИРОВАНИЕ СТРУКТУР
	// ========================================
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("[Эпоха 5] Тёмные века — формирование первых звёзд и галактик")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  T = 380 000 … 400 000 000 лет после Большого взрыва")
	fmt.Println()

	volumeRadius := OBSERVABLE_RADIUS * 0.001
	halos := StructureFormation(DARK_MATTER_DENSITY, volumeRadius)
	fmt.Printf("  Гало тёмной материи: %d\n", len(halos))
	totalDMMass := 0.0
	for _, h := range halos {
		totalDMMass += h.Mass
	}
	fmt.Printf("  Общая масса тёмной материи в симуляции: %.2e M☉\n", totalDMMass)
	fmt.Println()

	// ========================================
	// УРОВЕНЬ 6: ЗВЁЗДЫ, ПЛАНЕТЫ
	// ========================================
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("[Эпоха 6] Эпоха звёзд — формирование звёзд и планет")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  T = 400 000 000 … 13 800 000 000 лет")
	fmt.Println()

	allStars := make([]*Star, 0)
	allPlanets := make([]*Planet, 0)
	habitablePlanets := make([]*Planet, 0)

	for _, halo := range halos {
		stars := CreateStars(halo, rand.Intn(STARS_PER_GALAXY)+10)
		for _, star := range stars {
			planets := CreatePlanets(star)
			star.Planets = planets
			for _, planet := range planets {
				if planet.HasWater && planet.HasAtmosphere && planet.Temperature > 200 && planet.Temperature < 350 {
					habitablePlanets = append(habitablePlanets, planet)
				}
			}
			allPlanets = append(allPlanets, planets...)
		}
		allStars = append(allStars, stars...)
	}

	fmt.Printf("  Всего звёзд в симуляции: %d\n", len(allStars))
	fmt.Printf("  Всего планет: %d\n", len(allPlanets))
	fmt.Printf("  Планет в зоне обитаемости: %d\n", len(habitablePlanets))
	fmt.Println()

	// Распределение спектральных классов
	spectralCounts := make(map[byte]int)
	for _, star := range allStars {
		spectralCounts[star.SpectralClass]++
	}
	fmt.Println("  Распределение звёзд по спектральным классам:")
	for _, class := range []byte{'O', 'B', 'A', 'F', 'G', 'K', 'M'} {
		if cnt, ok := spectralCounts[class]; ok {
			fmt.Printf("    Класс %c: %d (%.1f%%)\n", class, cnt,
				float64(cnt)/float64(len(allStars))*100)
		}
	}
	fmt.Println()

	// ========================================
	// УРОВЕНЬ 7: АБИОГЕНЕЗ
	// ========================================
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("[Эпоха 7] Абиогенез — возникновение жизни")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	lifeFound := false
	var lifePlanet *Planet
	var lifeStar *Star

	for _, planet := range habitablePlanets {
		if Abiogenesis(planet) {
			lifeFound = true
			lifePlanet = planet
			lifeStar = planet.HostStar
			break
		}
	}

	if !lifeFound {
		fmt.Println("  ✦ ЖИЗНЬ НЕ ВОЗНИКЛА В ЭТОЙ ВСЕЛЕННОЙ ✦")
		fmt.Println("  Причина: недостаточно подходящих планет или не повезло")
		fmt.Println("  Перезапустите симуляцию для другой реализации")
		fmt.Println()
		fmt.Println("╔══════════════════════════════════════════════════════════════╗")
		fmt.Println("║  СИМУЛЯЦИЯ ЗАВЕРШЕНА: БЕЗЖИЗНЕННАЯ ВСЕЛЕННАЯ               ║")
		fmt.Println("╚══════════════════════════════════════════════════════════════╝")
		return
	}

	fmt.Printf("  ✦ ЖИЗНЬ ВОЗНИКЛА! ✦\n")
	fmt.Printf("  Планета:\n")
	fmt.Printf("    Температура: %.0f K\n", lifePlanet.Temperature)
	fmt.Printf("    Вода: %v\n", lifePlanet.HasWater)
	fmt.Printf("    Атмосфера: %v\n", lifePlanet.HasAtmosphere)
	fmt.Printf("    Магнитное поле: %v\n", lifePlanet.HasMagneticField)
	fmt.Printf("    Тип: %s / %s\n", lifePlanet.Composition, lifePlanet.Type)
	fmt.Printf("  Родительская звезда:\n")
	fmt.Printf("    Спектральный класс: %c%s\n", lifeStar.SpectralClass, lifeStar.LuminosityClass)
	fmt.Printf("    Масса: %.2f M☉\n", lifeStar.Mass)
	fmt.Printf("    Температура: %.0f K\n", lifeStar.Temperature)
	fmt.Printf("    Возраст: %.1f млрд лет\n", lifeStar.Age)
	fmt.Println()

	// ========================================
	// УРОВЕНЬ 8: БИОЛОГИЧЕСКАЯ ЭВОЛЮЦИЯ
	// ========================================
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("[Эпоха 8] Биологическая эволюция")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  T = 500 000 000 … 4 500 000 000 лет после возникновения жизни")
	fmt.Println()

	organism := CreateLUCA()
	fmt.Printf("  Начальный организм: %s\n", organism.Species)
	fmt.Printf("  Геном: %d млн п.н.\n", organism.GenomeSize)
	fmt.Println()

	// Эволюция на протяжении миллиарда поколений
	totalGenerations := 1000000000 // 1 млрд поколений (~4 млрд лет / 4 года на поколение)
	reportInterval := 100000000

	for gen := int64(0); gen < int64(totalGenerations); gen++ {
		organism = Evolve(organism, *lifePlanet)
		organism.Generation = gen

		if gen%int64(reportInterval) == 0 {
			fmt.Printf("  Поколение %d: вид %s, сложность %.3f, интеллект %.3f, популяция %d\n",
				gen, organism.Species, organism.Complexity, organism.Intelligence, organism.Population)
		}

		if organism.IsExtinct {
			fmt.Printf("  ✦ ВИД ВЫМЕР НА ПОКОЛЕНИИ %d ✦\n", gen)
			fmt.Println("  Перезапустите симуляцию для другой эволюционной траектории")
			return
		}

		// Появление разума
		if organism.Intelligence > 0.9 && organism.Species == "Homo Sapiens" {
			fmt.Printf("  ✦ РАЗУМНАЯ ЖИЗНЬ ПОЯВИЛАСЬ НА ПОКОЛЕНИИ %d ✦\n", gen)
			break
		}
	}

	fmt.Printf("  Финальный вид: %s\n", organism.Species)
	fmt.Printf("  Сложность: %.4f\n", organism.Complexity)
	fmt.Printf("  Интеллект: %.4f\n", organism.Intelligence)
	fmt.Printf("  Популяция: %d\n", organism.Population)
	fmt.Printf("  Объём мозга: %.0f см³\n", organism.BrainSize)
	fmt.Printf("  Масса тела: %.1f кг\n", organism.BodyMass)
	fmt.Println()

	// ========================================
	// УРОВЕНЬ 9: ТЕХНОЛОГИЧЕСКАЯ ЦИВИЛИЗАЦИЯ
	// ========================================
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("[Эпоха 9] Технологическая цивилизация")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	civilization := DevelopCivilization(organism, lifePlanet)
	fmt.Printf("  Технологический уровень: %.2f\n", civilization.TechLevel)
	fmt.Printf("  Население: %d\n", civilization.Population)
	fmt.Printf("  Прошла Великий фильтр: %v\n", civilization.SurvivedFilter)
	fmt.Printf("  Шкала Кардашёва: %.2f\n", civilization.KardashevScale)
	fmt.Printf("  Компьютеры: %v\n", civilization.HasComputers)
	fmt.Printf("  Интернет: %v\n", civilization.HasInternet)
	fmt.Printf("  Искусственный интеллект: %v\n", civilization.HasAI)
	fmt.Printf("  Язык Go: %v\n", civilization.HasGo)
	fmt.Printf("  Go-программистов: %d\n", len(civilization.GoProgrammers))
	fmt.Println()

	// Технологическая история
	fmt.Println("  Ключевые изобретения:")
	for _, tech := range civilization.Technologies {
		if tech.Impact > 0.5 {
			fmt.Printf("    Год %d: %s (влияние: %.1f)\n", tech.Year, tech.Name, tech.Impact)
		}
	}
	fmt.Println()

	// ========================================
	// УРОВЕНЬ 10: HELLO WORLD
	// ========================================
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("[Эпоха 10] Hello World — цель Вселенной")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	if !civilization.HasGo {
		fmt.Println("  Цивилизация не создала язык Go.")
		fmt.Println("  Возможно, они используют Rust. Или Python. Или что-то совсем иное.")
		fmt.Println("  Симуляция завершена без достижения цели.")
		return
	}

	program := WriteHelloWorld(civilization)
	if program == nil {
		fmt.Println("  Не удалось написать программу.")
		return
	}

	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("  ║            ФИНАЛЬНЫЙ РЕЗУЛЬТАТ СИМУЛЯЦИИ                   ║")
	fmt.Println("  ╠══════════════════════════════════════════════════════════════╣")
	fmt.Printf("  ║  Вселенная:       %.0f млрд лет эволюции                  ║\n", UNIVERSE_AGE/1e9)
	fmt.Printf("  ║  Автор:           %-40s ║\n", program.Author)
	fmt.Println("  ╠══════════════════════════════════════════════════════════════╣")
	fmt.Println("  ║  Исходный код программы:                                   ║")
	fmt.Println("  ║                                                            ║")
	fmt.Println("  ║    package main                                            ║")
	fmt.Println("  ║                                                            ║")
	fmt.Println("  ║    import \"fmt\"                                            ║")
	fmt.Println("  ║                                                            ║")
	fmt.Println("  ║    func main() {                                           ║")
	fmt.Println("  ║        fmt.Println(\"hello world\")                           ║")
	fmt.Println("  ║    }                                                       ║")
	fmt.Println("  ║                                                            ║")
	fmt.Println("  ╠══════════════════════════════════════════════════════════════╣")
	fmt.Printf("  ║  Вывод:           %-40s ║\n", program.Output)
	fmt.Println("  ╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("  ✦ СИМУЛЯЦИЯ УСПЕШНО ЗАВЕРШЕНА ✦")
	fmt.Println()
	fmt.Printf("  От квантового вакуума до \"%s\" —\n", program.Output)
	fmt.Println("  вся история Вселенной в 13.8 миллиардах лет.")
	fmt.Println()

	// Вывод напрямую в видеопамять (bare metal)
	vgaPrintHelloWorld()

	// Симуляция завершена. Процесс продолжает висеть в памяти,
	// сохраняя результат — подобно тому, как информация о нашей
	// Вселенной сохраняется на её космологическом горизонте.
	select {}
}
