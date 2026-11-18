import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { ReactNode, TouchEvent } from 'react'
import { BrowserRouter, Link, Route, Routes, useNavigate } from 'react-router-dom'
import {
  ANIMALS,
  BASE,
  COLORS,
  PHOTO_CREDITS,
  SOUND_CREDITS,
  ambientNote,
  letterUrl,
  photoUrl,
  shuffle,
  soundUrl,
  voiceUrl,
} from './data'
import type { Animal, Credit, GameColor } from './data'
import { playClip, playSequence, stopAudio } from './audio'
import './game.css'

/** Public, login-free kids' game served at /interactive-game. Rendered
 *  by main.tsx instead of the finance app, outside AuthProvider. */
export function GameApp() {
  useEffect(() => {
    document.title = 'Tebak-tebakan'
    document.body.classList.add('kg-body')
  }, [])

  return (
    <BrowserRouter basename={BASE}>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/hewan" element={<AnimalCards />} />
        <Route path="/hewan/kuis" element={<Quiz kind="animal" />} />
        <Route path="/warna" element={<ColorCards />} />
        <Route path="/warna/kuis" element={<Quiz kind="color" />} />
        <Route path="/sumber" element={<Credits />} />
        <Route path="*" element={<Home />} />
      </Routes>
    </BrowserRouter>
  )
}

// ─── Shared bits ────────────────────────────────────────────────────────

function Screen({ children, tone }: { children: ReactNode; tone: 'home' | 'animal' | 'color' }) {
  useEffect(() => stopAudio, [])
  return <div className={`kg kg-${tone}`}>{children}</div>
}

function TopBar({ title, section, mode }: { title: string; section?: 'hewan' | 'warna'; mode?: 'kartu' | 'kuis' }) {
  const navigate = useNavigate()
  return (
    <header className="kg-top">
      <button className="kg-back" onClick={() => navigate('/')} aria-label="Kembali ke menu">
        ←
      </button>
      <h1>{title}</h1>
      {section && (
        <nav className="kg-modes" aria-label="Mode">
          <Link className={mode === 'kartu' ? 'on' : ''} to={`/${section}`} replace>
            Kartu
          </Link>
          <Link className={mode === 'kuis' ? 'on' : ''} to={`/${section}/kuis`} replace>
            Kuis
          </Link>
        </nav>
      )}
    </header>
  )
}

function BigButton({
  icon,
  label,
  onClick,
  active,
  color,
}: {
  icon: string
  label: string
  onClick: () => void
  active?: boolean
  color: 'orange' | 'blue' | 'green' | 'pink'
}) {
  return (
    <button className={`kg-btn kg-btn-${color}${active ? ' playing' : ''}`} onClick={onClick}>
      <span className="kg-btn-icon" aria-hidden>
        {icon}
      </span>
      {label}
    </button>
  )
}

/** Left/right swipe on a card. */
function useSwipe(onPrev: () => void, onNext: () => void) {
  const start = useRef<{ x: number; y: number } | null>(null)
  return {
    onTouchStart: (e: TouchEvent) => {
      const t = e.touches[0]
      start.current = { x: t.clientX, y: t.clientY }
    },
    onTouchEnd: (e: TouchEvent) => {
      const s = start.current
      start.current = null
      if (!s) return
      const t = e.changedTouches[0]
      const dx = t.clientX - s.x
      const dy = t.clientY - s.y
      if (Math.abs(dx) > 50 && Math.abs(dx) > Math.abs(dy) * 1.5) {
        if (dx < 0) onNext()
        else onPrev()
      }
    },
  }
}

/** Runs a clip and reports which button is "playing" so it can glow. */
function usePlayer() {
  const [playing, setPlaying] = useState<string | null>(null)
  const run = useCallback(async (key: string, task: () => Promise<boolean>) => {
    setPlaying(key)
    const finished = await task()
    if (finished) setPlaying(null)
  }, [])
  const reset = useCallback(() => {
    stopAudio()
    setPlaying(null)
  }, [])
  return { playing, run, reset }
}

function CardNav({
  idx,
  total,
  onPrev,
  onNext,
  onShuffle,
  onGrid,
}: {
  idx: number
  total: number
  onPrev: () => void
  onNext: () => void
  onShuffle: () => void
  onGrid: () => void
}) {
  return (
    <div className="kg-nav">
      <button className="kg-round" onClick={onShuffle} aria-label="Acak urutan">
        🔀
      </button>
      <button className="kg-arrow" onClick={onPrev} aria-label="Sebelumnya">
        ◀
      </button>
      <span className="kg-count">
        {idx + 1} / {total}
      </span>
      <button className="kg-arrow" onClick={onNext} aria-label="Berikutnya">
        ▶
      </button>
      <button className="kg-round" onClick={onGrid} aria-label="Pilih dari semua">
        ▦
      </button>
    </div>
  )
}

function useDeck<T extends { id: string }>(items: T[]) {
  const [order, setOrder] = useState(items)
  const [idx, setIdx] = useState(0)
  const go = useCallback((d: number) => setIdx((i) => (i + d + order.length) % order.length), [order.length])
  const reshuffle = useCallback(() => {
    setOrder(shuffle(items))
    setIdx(0)
  }, [items])
  const jumpTo = useCallback((id: string) => setIdx(Math.max(0, order.findIndex((x) => x.id === id))), [order])
  return { order, idx, item: order[idx], go, reshuffle, jumpTo }
}

// ─── Home ───────────────────────────────────────────────────────────────

function Home() {
  return (
    <Screen tone="home">
      <div className="kg-home">
        <div className="kg-hello">
          <span className="kg-wave" aria-hidden>
            👋
          </span>
          <h1>
            Ayo main <span className="kg-nowrap">tebak-tebakan!</span>
          </h1>
        </div>
        <Link to="/hewan" className="kg-tile kg-tile-animal">
          <img src={photoUrl('cat')} alt="" />
          <span>Tebak Hewan</span>
        </Link>
        <Link to="/warna" className="kg-tile kg-tile-color">
          <span className="kg-dots" aria-hidden>
            {COLORS.map((c) => (
              <i key={c.id} style={{ background: c.hex }} />
            ))}
          </span>
          <span>Tebak Warna</span>
        </Link>
        <Link to="/sumber" className="kg-credits-link">
          Sumber foto & suara
        </Link>
      </div>
    </Screen>
  )
}

// ─── Flash cards ────────────────────────────────────────────────────────

function AnimalCards() {
  const deck = useDeck(ANIMALS)
  const { playing, run, reset } = usePlayer()
  const [revealed, setRevealed] = useState(false)
  const [grid, setGrid] = useState(false)
  const a = deck.item

  useEffect(() => {
    reset()
    setRevealed(false)
    // Warm the next photo so swiping feels instant.
    const next = deck.order[(deck.idx + 1) % deck.order.length]
    new Image().src = photoUrl(next.id)
  }, [a.id, deck.idx, deck.order, reset])

  const swipe = useSwipe(() => deck.go(-1), () => deck.go(1))
  const note = ambientNote(a.id)

  return (
    <Screen tone="animal">
      <TopBar title="Tebak Hewan" section="hewan" mode="kartu" />
      <main className="kg-card-wrap" {...swipe}>
        <figure className="kg-photo" onClick={() => setRevealed(true)}>
          <Photo key={a.id} id={a.id} />
        </figure>
        <div className="kg-answer">
          {revealed ? (
            <div className="kg-name pop">{a.name}</div>
          ) : (
            <button className="kg-reveal" onClick={() => setRevealed(true)}>
              Ini hewan apa? <b>Lihat jawaban 👀</b>
            </button>
          )}
          {playing === 'sound' && note && <div className="kg-note">{note}</div>}
        </div>
        <div className="kg-actions">
          <BigButton
            icon="🔊"
            label="Suara"
            color="orange"
            active={playing === 'sound'}
            onClick={() => run('sound', () => playClip(soundUrl(a.id)))}
          />
          <BigButton
            icon="🗣️"
            label="Baca nama"
            color="blue"
            active={playing === 'name'}
            onClick={() => {
              setRevealed(true)
              run('name', () => playClip(voiceUrl(`name-${a.id}`)))
            }}
          />
        </div>
        <CardNav
          idx={deck.idx}
          total={deck.order.length}
          onPrev={() => deck.go(-1)}
          onNext={() => deck.go(1)}
          onShuffle={deck.reshuffle}
          onGrid={() => setGrid(true)}
        />
      </main>
      {grid && (
        <Picker onClose={() => setGrid(false)}>
          {deck.order.map((x) => (
            <button
              key={x.id}
              className="kg-pick-photo"
              onClick={() => {
                deck.jumpTo(x.id)
                setGrid(false)
              }}
            >
              <img src={photoUrl(x.id)} alt="" loading="lazy" />
              <span>{x.name}</span>
            </button>
          ))}
        </Picker>
      )}
    </Screen>
  )
}

function ColorCards() {
  const deck = useDeck(COLORS)
  const { playing, run, reset } = usePlayer()
  const [revealed, setRevealed] = useState(false)
  const [lit, setLit] = useState<number | null>(null)
  const [grid, setGrid] = useState(false)
  const c = deck.item

  useEffect(() => {
    reset()
    setRevealed(false)
    setLit(null)
  }, [c.id, reset])

  const swipe = useSwipe(() => deck.go(-1), () => deck.go(1))
  const letters = c.name.toUpperCase().split('')

  const spell = () => {
    setRevealed(true)
    // Letter by letter ("ka, u, en, i, en, ge") then the whole word; the
    // hyphen in Abu-abu becomes a short pause.
    const urls = [...letters.map((ch) => (ch === '-' ? null : letterUrl(ch))), voiceUrl(`name-${c.id}`)]
    run('spell', async () => {
      const ok = await playSequence(urls, (i) => setLit(i < letters.length ? i : -1), 120)
      setLit(null)
      return ok
    })
  }

  return (
    <Screen tone="color">
      <TopBar title="Tebak Warna" section="warna" mode="kartu" />
      <main className="kg-card-wrap" {...swipe}>
        <figure className="kg-swatch-card" onClick={() => setRevealed(true)}>
          <Swatch color={c} big />
        </figure>
        <div className="kg-answer">
          {revealed ? (
            <div className={`kg-letters${lit === -1 ? ' whole' : ''}`} aria-label={c.name}>
              {letters.map((ch, i) => (
                <span key={i} className={ch === '-' ? 'dash' : lit === i ? 'lit' : ''}>
                  {ch}
                </span>
              ))}
            </div>
          ) : (
            <button className="kg-reveal" onClick={() => setRevealed(true)}>
              Ini warna apa? <b>Lihat jawaban 👀</b>
            </button>
          )}
        </div>
        <div className="kg-actions">
          <BigButton
            icon="🗣️"
            label="Baca nama"
            color="blue"
            active={playing === 'name'}
            onClick={() => {
              setRevealed(true)
              run('name', () => playClip(voiceUrl(`name-${c.id}`)))
            }}
          />
          <BigButton icon="🔤" label="Eja" color="green" active={playing === 'spell'} onClick={spell} />
        </div>
        <CardNav
          idx={deck.idx}
          total={deck.order.length}
          onPrev={() => deck.go(-1)}
          onNext={() => deck.go(1)}
          onShuffle={deck.reshuffle}
          onGrid={() => setGrid(true)}
        />
      </main>
      {grid && (
        <Picker onClose={() => setGrid(false)}>
          {deck.order.map((x) => (
            <button
              key={x.id}
              className="kg-pick-color"
              onClick={() => {
                deck.jumpTo(x.id)
                setGrid(false)
              }}
            >
              <Swatch color={x} />
              <span>{x.name}</span>
            </button>
          ))}
        </Picker>
      )}
    </Screen>
  )
}

/** The whole photo, never cropped (heads and tails matter when guessing),
 *  over a blurred copy of itself so the frame isn't left with bars. */
function Photo({ id }: { id: string }) {
  const src = photoUrl(id)
  return (
    <span className="kg-photo-fit">
      <img className="kg-photo-bg" src={src} alt="" aria-hidden draggable={false} />
      <img className="kg-photo-fg" src={src} alt="Foto hewan" draggable={false} />
    </span>
  )
}

function Swatch({ color, big }: { color: GameColor; big?: boolean }) {
  return (
    <span
      className={`kg-swatch${big ? ' big' : ''}${color.id === 'putih' ? ' light' : ''}`}
      style={{ background: color.hex }}
    />
  )
}

function Picker({ children, onClose }: { children: ReactNode; onClose: () => void }) {
  return (
    <div className="kg-sheet" role="dialog" aria-modal onClick={onClose}>
      <div className="kg-sheet-inner" onClick={(e) => e.stopPropagation()}>
        <div className="kg-sheet-head">
          <b>Pilih</b>
          <button className="kg-round" onClick={onClose} aria-label="Tutup">
            ✕
          </button>
        </div>
        <div className="kg-sheet-grid">{children}</div>
      </div>
    </div>
  )
}

// ─── Quiz ───────────────────────────────────────────────────────────────

type QuizItem = (Animal | GameColor) & { id: string; name: string }

const OPTIONS = 4

function Quiz({ kind }: { kind: 'animal' | 'color' }) {
  const items: QuizItem[] = kind === 'animal' ? ANIMALS : COLORS
  // Targets come from a shuffled deck so every item gets asked before any
  // repeats.
  const deckRef = useRef<QuizItem[]>([])
  const nextTarget = useCallback(() => {
    if (deckRef.current.length === 0) deckRef.current = shuffle(items)
    return deckRef.current.pop()!
  }, [items])

  const makeRound = useCallback(() => {
    const target = nextTarget()
    const others = shuffle(items.filter((x) => x.id !== target.id)).slice(0, OPTIONS - 1)
    return { target, options: shuffle([target, ...others]), key: Math.random() }
  }, [items, nextTarget])

  const [round, setRound] = useState(makeRound)
  const [wrong, setWrong] = useState<Set<string>>(new Set())
  const [solved, setSolved] = useState(false)
  const [stars, setStars] = useState(0)
  const [started, setStarted] = useState(false)

  const ask = useCallback(() => playClip(voiceUrl(`ask-${round.target.id}`)), [round.target.id])

  // Read each new question aloud — but only once the player has tapped
  // "Mulai", since phones refuse to play sound before the first tap.
  useEffect(() => {
    if (!started) return
    for (const o of round.options) if (kind === 'animal') new Image().src = photoUrl(o.id)
    ask()
  }, [round.key, started]) // eslint-disable-line react-hooks/exhaustive-deps

  const pick = async (o: QuizItem) => {
    if (solved) return
    if (o.id !== round.target.id) {
      setWrong((w) => new Set(w).add(o.id))
      playClip(voiceUrl('no-1'))
      return
    }
    setSolved(true)
    setStars((s) => s + 1)
    const praise = voiceUrl(`ok-${1 + Math.floor(Math.random() * 3)}`)
    const seq = [praise, voiceUrl(`name-${o.id}`)]
    if (kind === 'animal' && !ambientNote(o.id)) seq.push(soundUrl(o.id))
    const finished = await playSequence(seq, undefined, 150)
    if (finished) {
      await new Promise((r) => setTimeout(r, 500))
      nextRound()
    }
  }

  const nextRound = () => {
    stopAudio()
    setWrong(new Set())
    setSolved(false)
    setRound(makeRound())
  }

  const title = kind === 'animal' ? 'Tebak Hewan' : 'Tebak Warna'
  const section = kind === 'animal' ? 'hewan' : 'warna'
  const promptName = round.target.name.toLowerCase()

  return (
    <Screen tone={kind}>
      <TopBar title={title} section={section} mode="kuis" />
      <main className="kg-quiz">
        {!started ? (
          <div className="kg-start">
            <p>Dengarkan pertanyaannya, lalu pilih jawaban yang benar!</p>
            <button className="kg-btn kg-btn-green kg-start-btn" onClick={() => setStarted(true)}>
              <span className="kg-btn-icon" aria-hidden>
                ▶
              </span>
              Mulai
            </button>
          </div>
        ) : (
          <>
            <div className="kg-question">
              <button className="kg-round kg-replay" onClick={ask} aria-label="Ulangi pertanyaan">
                🔊
              </button>
              <h2>
                Mana {kind === 'color' ? 'warna ' : ''}
                <b>{promptName}</b>?
              </h2>
              <span className="kg-stars" aria-label={`${stars} bintang`}>
                ⭐ {stars}
              </span>
            </div>
            <div className={`kg-options kg-options-${kind}`} key={round.key}>
              {round.options.map((o) => {
                const isWrong = wrong.has(o.id)
                const isRight = solved && o.id === round.target.id
                return (
                  <button
                    key={o.id}
                    className={`kg-option${isWrong ? ' wrong' : ''}${isRight ? ' right' : ''}${solved && !isRight ? ' faded' : ''}`}
                    onClick={() => pick(o)}
                    aria-label={kind === 'color' ? `Pilihan warna` : 'Pilihan hewan'}
                  >
                    {kind === 'animal' ? (
                      <Photo id={o.id} />
                    ) : (
                      <Swatch color={o as GameColor} />
                    )}
                    {isRight && <span className="kg-check">✓</span>}
                  </button>
                )
              })}
            </div>
            {solved && (
              <button className="kg-next" onClick={nextRound}>
                Lanjut ▶
              </button>
            )}
          </>
        )}
      </main>
    </Screen>
  )
}

// ─── Credits ────────────────────────────────────────────────────────────

function CreditLine({ label, c }: { label: string; c?: Credit }) {
  if (!c) return null
  return (
    <span className="kg-credit-line">
      {label}:{' '}
      <a href={c.sourceUrl} target="_blank" rel="noreferrer">
        {c.author || c.title.replace(/^File:/, '')}
      </a>{' '}
      ·{' '}
      {c.licenseUrl ? (
        <a href={c.licenseUrl} target="_blank" rel="noreferrer">
          {c.license}
        </a>
      ) : (
        c.license
      )}
    </span>
  )
}

function Credits() {
  const rows = useMemo(() => ANIMALS.map((a) => ({ a, photo: PHOTO_CREDITS[a.id], sound: SOUND_CREDITS[a.id] })), [])
  return (
    <Screen tone="home">
      <header className="kg-top">
        <Link className="kg-back" to="/" aria-label="Kembali ke menu">
          ←
        </Link>
        <h1>Sumber foto & suara</h1>
      </header>
      <main className="kg-credits">
        <p>
          Semua foto dan suara hewan berasal dari Wikimedia Commons, dipakai sesuai lisensinya. Suara pembaca nama
          dibuat dengan suara komputer berbahasa Indonesia.
        </p>
        <ul>
          {rows.map(({ a, photo, sound }) => (
            <li key={a.id}>
              <img src={photoUrl(a.id)} alt="" loading="lazy" />
              <div>
                <b>{a.name}</b>
                <CreditLine label="Foto" c={photo} />
                <CreditLine label={sound?.kind === 'ambient' ? 'Suara (suasana)' : 'Suara'} c={sound} />
              </div>
            </li>
          ))}
        </ul>
      </main>
    </Screen>
  )
}
