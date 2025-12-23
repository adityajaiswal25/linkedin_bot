package stealth

import (
	"math"
	"math/rand"
	"time"

	"linkedin-automation/pkg/config"

	"github.com/go-rod/rod"
)

// Stealth implements anti-bot detection techniques
type Stealth struct {
	cfg  *config.Config
	page *rod.Page
	rng  *rand.Rand
}

// NewStealth creates a new stealth instance
func NewStealth(cfg *config.Config, page *rod.Page) *Stealth {
	return &Stealth{
		cfg:  cfg,
		page: page,
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Apply applies all enabled stealth techniques
func (s *Stealth) Apply() error {
	if s.cfg.Stealth.Fingerprint.Enabled {
		s.applyFingerprintMasking()
	}

	if s.cfg.Stealth.MouseMovement.Enabled {
		// mouse movement handled in helpers
	}

	if s.cfg.Stealth.Timing.Enabled {
		s.applyTiming()
	}

	return nil
}

// ShouldOperate checks if operations should proceed based on scheduling
func (s *Stealth) ShouldOperate() bool {
	if !s.cfg.Stealth.Scheduling.Enabled || !s.cfg.Stealth.Scheduling.BusinessHoursOnly {
		return true
	}

	now := time.Now()
	hour := now.Hour()

	return hour >= s.cfg.Stealth.Scheduling.StartHour && hour < s.cfg.Stealth.Scheduling.EndHour
}

// RandomBreak takes a random break
func (s *Stealth) RandomBreak() {
	if !s.cfg.Stealth.Scheduling.Enabled {
		return
	}

	if s.rng.Float64() < s.cfg.Stealth.Scheduling.BreakProbability {
		duration := time.Duration(s.rng.Intn(300)+60) * time.Second // 1-5 minutes
		time.Sleep(duration)
	}
}

// HumanClick performs a human-like click
func (s *Stealth) HumanClick(el *rod.Element) {
	if s.cfg.Stealth.MouseMovement.Enabled {
		s.moveMouseToElement(el)
	}

	if s.cfg.Stealth.Hovering.Enabled && s.rng.Float64() < s.cfg.Stealth.Hovering.HoverProbability {
		s.hoverOverElement(el)
	}

	el.MustClick()
}

// HumanType performs human-like typing
func (s *Stealth) HumanType(el *rod.Element, text string) {
	el.MustFocus()

	for _, char := range text {
		el.MustInput(string(char))

		if s.cfg.Stealth.Typing.Enabled {
			delay := time.Duration(s.rng.Intn(s.cfg.Stealth.Typing.MaxKeystrokeDelay-s.cfg.Stealth.Typing.MinKeystrokeDelay)+s.cfg.Stealth.Typing.MinKeystrokeDelay) * time.Millisecond

			// Occasional typo
			if s.rng.Float64() < s.cfg.Stealth.Typing.TypoProbability {
				el.MustInput("x") // wrong character
				time.Sleep(delay)
				el.MustInput("\b") // backspace
				time.Sleep(delay)
			}

			time.Sleep(delay)
		}
	}
}

// ScrollHumanLike performs human-like scrolling
func (s *Stealth) ScrollHumanLike(distance int) {
	if !s.cfg.Stealth.Scrolling.Enabled {
		s.page.MustEval("window.scrollBy(0, ?)", distance)
		return
	}

	steps := s.rng.Intn(10) + 5 // 5-15 steps
	stepSize := float64(distance) / float64(steps)

	for i := 0; i < steps; i++ {
		variance := 1.0 + (s.rng.Float64()-0.5)*s.cfg.Stealth.Timing.ScrollSpeedVariance
		actualStep := int(stepSize * variance)

		s.page.MustEval("window.scrollBy(0, ?)", actualStep)

		delay := time.Duration(s.rng.Intn(200)+50) * time.Millisecond
		time.Sleep(delay)
	}

	// Occasional scroll back
	if s.rng.Float64() < s.cfg.Stealth.Scrolling.ScrollBackProbability {
		backDistance := s.rng.Intn(distance/4) + 10
		s.page.MustEval("window.scrollBy(0, ?)", -backDistance)
		time.Sleep(time.Duration(s.rng.Intn(1000)+500) * time.Millisecond)
	}
}

// RandomDelay adds a random delay
func (s *Stealth) RandomDelay() {
	if !s.cfg.Stealth.Timing.Enabled {
		return
	}

	min := s.cfg.Stealth.Timing.MinThinkTime
	max := s.cfg.Stealth.Timing.MaxThinkTime
	if max <= min {
		max = min + 1
	}
	delay := time.Duration(s.rng.Intn(max-min)+min) * time.Millisecond
	time.Sleep(delay)
}

func (s *Stealth) applyFingerprintMasking() {
	// Randomize viewport
	if s.cfg.Stealth.Fingerprint.RandomizeViewport {
		width := s.cfg.Browser.Viewport.Width + rand.Intn(100) - 50
		height := s.cfg.Browser.Viewport.Height + rand.Intn(100) - 50
		s.page.MustSetViewport(width, height, 1.0, false)
	}

	// Disable webdriver flag
	if s.cfg.Stealth.Fingerprint.DisableWebdriverFlag {
		s.page.MustEval("Object.defineProperty(navigator, 'webdriver', {get: () => undefined})")
	}
}

func (s *Stealth) applyMouseMovement() {
	// Mouse movement is handled via moveMouseToElement / HumanClick helpers.
}

func (s *Stealth) applyTiming() {
	// Add random delays to actions
}

func (s *Stealth) moveMouseToElement(el *rod.Element) {
	box := el.MustBox()
	targetX := box.X + box.Width/2 + (s.rng.Float64()-0.5)*20
	targetY := box.Y + box.Height/2 + (s.rng.Float64()-0.5)*20

	// Approximate current mouse position as viewport center
	vp := s.page.MustEval(`() => ({ w: window.innerWidth, h: window.innerHeight })`)
	currentX := vp.Get("w").Float() / 2
	currentY := vp.Get("h").Float() / 2

	// Generate cubic Bézier control points
	cp1X := currentX + (targetX-currentX)*0.3 + (s.rng.Float64()-0.5)*30
	cp1Y := currentY + (targetY-currentY)*0.3 + (s.rng.Float64()-0.5)*20
	cp2X := currentX + (targetX-currentX)*0.6 + (s.rng.Float64()-0.5)*30
	cp2Y := currentY + (targetY-currentY)*0.6 + (s.rng.Float64()-0.5)*20

	steps := s.rng.Intn(15) + 25 // 25-40 steps
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := bezierCurve(t, currentX, cp1X, cp2X, targetX)
		y := bezierCurve(t, currentY, cp1Y, cp2Y, targetY)

		// Micro-corrections / jitter
		if s.cfg.Stealth.MouseMovement.MicroCorrections && s.rng.Float64() < 0.1 {
			x += (s.rng.Float64() - 0.5) * 2
			y += (s.rng.Float64() - 0.5) * 2
		}

		s.page.Mouse.Move(x, y, 0)

		// Variable speed easing
		ease := t * t * (3 - 2*t)
		base := 4 + s.rng.Intn(6) // 4-9 ms
		sleep := time.Duration(float64(base) * (0.5 + ease*0.8) * float64(time.Millisecond))
		time.Sleep(sleep)
	}
}

func (s *Stealth) hoverOverElement(el *rod.Element) {
	duration := time.Duration(s.rng.Intn(s.cfg.Stealth.Hovering.HoverDurationMax-s.cfg.Stealth.Hovering.HoverDurationMin)+s.cfg.Stealth.Hovering.HoverDurationMin) * time.Millisecond
	time.Sleep(duration)
}

// bezierCurve calculates a point on a cubic Bézier curve
func bezierCurve(t float64, p0, p1, p2, p3 float64) float64 {
	return math.Pow(1-t, 3)*p0 + 3*math.Pow(1-t, 2)*t*p1 + 3*(1-t)*math.Pow(t, 2)*p2 + math.Pow(t, 3)*p3
}
