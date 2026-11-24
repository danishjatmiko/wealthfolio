import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { ReactNode, TouchEvent } from 'react'
import { BrowserRouter, Link, Route, Routes, useNavigate } from 'react-router-dom'
import {
  ANIMALS,
  BASE,
  COLORS,
  EMOTIONS,
  FRUITS,
  PHOTO_CREDITS,
  SOUND_CREDITS,
  ambientNote,
  letterUrl,
  photoCount,
  photoUrl,
  randomPhoto,
  shuffle,
  soundUrl,
  voiceUrl,
} from './data'
import type { Credit, GameColor, Item, PhotoKind } from './data'
import { playClip, playSequence, stopAudio } from './audio'
import './game.css'

type Kind = PhotoKind | 'color'

const SECTIONS: Record<Kind, { path: string; title: string; items: Item[]; question: string; quizPrompt: string }> = {
  animal: { path: 'hewan', title: 'Tebak Hewan', items: ANIMALS, question: 'Ini hewan apa?', quizPrompt: 'Mana' },
  fruit: { path: 'buah', title: 'Tebak Buah', items: FRUITS, question: 'Ini buah apa?', quizPrompt: 'Mana' },
  color: { path: 'warna', title: 'Tebak Warna', items: COLORS, question: 'Ini warna apa?', quizPrompt: 'Mana warna' },
  emotion: {
    path: 'emosi',
    title: 'Tebak Emosi',
    items: EMOTIONS,
    question: 'Dia sedang merasa apa?',
    quizPrompt: 'Mana yang sedang',
  },
}

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
        {(Object.keys(SECTIONS) as Kind[]).flatMap((k) => [
          <Route key={k} path={`/${SECTIONS[k].path}`} element={<Cards key={k} kind={k} />} />,
          <Route key={`${k}-quiz`} path={`/${SECTIONS[k].path}/kuis`} element={<Quiz key={k} kind={k} />} />,
        ])}
        <Route path="/sumber" element={<Credits />} />
        <Route path="*" element={<Home />} />
      </Routes>
    </BrowserRouter>
  )
}

// ─── Shared bits ────────────────────────────────────────────────────────

function Screen({ children, tone }: { children: ReactNode; tone: 'home' | Kind }) {
  useEffect(() => stopAudio, [])
  return <div className={`kg kg-${tone}`}>{children}</div>
}

function TopBar({ kind, mode }: { kind: Kind; mode: 'kartu' | 'kuis' }) {
  const navigate = useNavigate()
  const { path, title } = SECTIONS[kind]
  return (
    <header className="kg-top">
      <button className="kg-back" onClick={() => navigate('/')} aria-label="Kembali ke menu">
        ←
      </button>
      <h1>{title}</h1>
      <nav className="kg-modes" aria-label="Mode">
        <Link className={mode === 'kartu' ? 'on' : ''} to={`/${path}`} replace>
          Kartu
        </Link>
        <Link className={mode === 'kuis' ? 'on' : ''} to={`/${path}/kuis`} replace>
          Kuis
        </Link>
      </nav>
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

function useDeck(items: Item[]) {
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

/** The whole photo, never cropped (heads and tails matter when guessing),
 *  over a blurred copy of itself so the frame isn't left with bars. */
function Photo({ kind, id, n }: { kind: PhotoKind; id: string; n: number }) {
  const src = photoUrl(kind, id, n)
  return (
    <span className="kg-photo-fit" key={src}>
      <img className="kg-photo-bg" src={src} alt="" aria-hidden draggable={false} />
      <img className="kg-photo-fg" src={src} alt={kind === 'emotion' ? 'Ilustrasi ekspresi' : 'Foto'} draggable={false} />
    </span>
  )
}

function Swatch({ color, big }: { color: GameColor; big?: boolean }) {
  return (
    <span
      className={`kg-swatch${big ? ' big' : ''}${color.id === 'putih' ? ' light' : ''}`}
      style={{ background: color.fill }}
    />
  )
}

const asColor = (item: Item) => item as GameColor

/** Letter tiles plus a speller that lights each one up as it's read
 *  ("ka, u, en, i, en, ge"), then the whole word. Spaces and hyphens
 *  ("Biru Muda", "Abu-abu") become a short pause. */
function useSpelling() {
  const [lit, setLit] = useState<number | null>(null)
  const spell = useCallback(async (item: Item) => {
    const letters = item.name.toUpperCase().split('')
    const urls = [...letters.map((ch) => (/[A-Z]/.test(ch) ? letterUrl(ch) : null)), voiceUrl(`name-${item.id}`)]
    const ok = await playSequence(urls, (i) => setLit(i < letters.length ? i : -1), 120)
    setLit(null)
    return ok
  }, [])
  const clear = useCallback(() => setLit(null), [])
  return { lit, spell, clear }
}

function LetterTiles({ name, lit }: { name: string; lit: number | null }) {
  const letters = name.toUpperCase().split('')
  return (
    <div className={`kg-letters${lit === -1 ? ' whole' : ''}${letters.length > 7 ? ' long' : ''}`} aria-label={name}>
      {letters.map((ch, i) => (
        <span key={i} className={/[A-Z]/.test(ch) ? (lit === i ? 'lit' : '') : 'dash'}>
          {ch === ' ' ? '' : ch}
        </span>
      ))}
    </div>
  )
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
        <Link to="/hewan" className="kg-tile kg-tile-photo kg-tile-animal">
          <img src={photoUrl('animal', 'cat', 1)} alt="" />
          <span>Tebak Hewan</span>
        </Link>
        <Link to="/buah" className="kg-tile kg-tile-photo kg-tile-fruit">
          <img src={photoUrl('fruit', 'stroberi', 1)} alt="" />
          <span>Tebak Buah</span>
        </Link>
        <Link to="/warna" className="kg-tile kg-tile-color">
          <span className="kg-dots" aria-hidden>
            {COLORS.map((c) => (
              <i key={c.id} style={{ background: c.fill }} />
            ))}
          </span>
          <span>Tebak Warna</span>
        </Link>
        <Link to="/emosi" className="kg-tile kg-tile-photo kg-tile-emotion">
          <img src={photoUrl('emotion', 'senang', 1)} alt="" />
          <span>Tebak Emosi</span>
        </Link>
        <Link to="/sumber" className="kg-credits-link">
          Sumber foto & suara
        </Link>
      </div>
    </Screen>
  )
}

// ─── Flash cards ────────────────────────────────────────────────────────

function Cards({ kind }: { kind: Kind }) {
  const section = SECTIONS[kind]
  const deck = useDeck(section.items)
  const { playing, run, reset } = usePlayer()
  const { lit, spell, clear } = useSpelling()
  const [revealed, setRevealed] = useState(false)
  const [grid, setGrid] = useState(false)
  const [photoN, setPhotoN] = useState(1)
  const item = deck.item
  const photoKind = kind === 'color' ? null : kind
  const imageCount = photoKind ? photoCount(photoKind, item.id) : 0

  useEffect(() => {
    reset()
    clear()
    setRevealed(false)
    if (!photoKind) return
    // Pick an image for this visit and warm the first image of the next card.
    setPhotoN(randomPhoto(photoKind, item.id))
    const next = deck.order[(deck.idx + 1) % deck.order.length]
    new Image().src = photoUrl(photoKind, next.id, 1)
  }, [item.id, deck.idx, deck.order, photoKind, reset, clear])

  const swipe = useSwipe(() => deck.go(-1), () => deck.go(1))
  const note = kind === 'animal' ? ambientNote(item.id) : null
  const showLetters = kind !== 'animal'

  const readName = () => {
    setRevealed(true)
    run('name', () => playClip(voiceUrl(`name-${item.id}`)))
  }
  const doSpell = () => {
    setRevealed(true)
    run('spell', () => spell(item))
  }

  return (
    <Screen tone={kind}>
      <TopBar kind={kind} mode="kartu" />
      <main className="kg-card-wrap" {...swipe}>
        {photoKind ? (
          <figure className="kg-photo" onClick={() => setRevealed(true)}>
            <Photo kind={photoKind} id={item.id} n={photoN} />
            {imageCount > 1 && (
              <button
                className="kg-round kg-other-photo"
                aria-label={kind === 'emotion' ? 'Ilustrasi lain' : 'Foto lain'}
                onClick={(e) => {
                  e.stopPropagation()
                  setPhotoN((n) => (n % imageCount) + 1)
                }}
              >
                🔄
              </button>
            )}
          </figure>
        ) : (
          <figure className="kg-swatch-card" onClick={() => setRevealed(true)}>
            <Swatch key={item.id} color={asColor(item)} big />
          </figure>
        )}
        <div className="kg-answer">
          {!revealed ? (
            <button className="kg-reveal" onClick={() => setRevealed(true)}>
              {section.question} <b>Lihat jawaban 👀</b>
            </button>
          ) : showLetters ? (
            <LetterTiles name={item.name} lit={lit} />
          ) : (
            <div className="kg-name pop">{item.name}</div>
          )}
          {playing === 'sound' && note && <div className="kg-note">{note}</div>}
        </div>
        <div className="kg-actions">
          {kind === 'animal' && SOUND_CREDITS[item.id] && (
            <BigButton
              icon="🔊"
              label="Suara"
              color="orange"
              active={playing === 'sound'}
              onClick={() => run('sound', () => playClip(soundUrl(item.id)))}
            />
          )}
          <BigButton icon="🗣️" label="Baca nama" color="blue" active={playing === 'name'} onClick={readName} />
          {kind !== 'animal' && (
            <BigButton icon="🔤" label="Eja" color="green" active={playing === 'spell'} onClick={doSpell} />
          )}
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
              className={photoKind ? 'kg-pick-photo' : 'kg-pick-color'}
              onClick={() => {
                deck.jumpTo(x.id)
                setGrid(false)
              }}
            >
              {photoKind ? (
                <img src={photoUrl(photoKind, x.id, 1)} alt="" loading="lazy" />
              ) : (
                <Swatch color={asColor(x)} />
              )}
              <span>{x.name}</span>
            </button>
          ))}
        </Picker>
      )}
    </Screen>
  )
}

// ─── Quiz ───────────────────────────────────────────────────────────────

const OPTIONS = 4

function Quiz({ kind }: { kind: Kind }) {
  const { items, quizPrompt } = SECTIONS[kind]
  const photoKind = kind === 'color' ? null : kind
  const optionLayout = kind === 'animal' || kind === 'fruit' ? 'photo' : kind
  // Targets come from a shuffled deck so every item gets asked before any
  // repeats.
  const deckRef = useRef<Item[]>([])
  const nextTarget = useCallback(() => {
    if (deckRef.current.length === 0) deckRef.current = shuffle(items)
    return deckRef.current.pop()!
  }, [items])

  const makeRound = useCallback(() => {
    const target = nextTarget()
    const others = shuffle(items.filter((x) => x.id !== target.id)).slice(0, OPTIONS - 1)
    const options = shuffle([target, ...others]).map((item) => ({
      item,
      n: photoKind ? randomPhoto(photoKind, item.id) : 1,
    }))
    return { target, options, key: Math.random() }
  }, [items, nextTarget, photoKind])

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
    if (photoKind) for (const o of round.options) new Image().src = photoUrl(photoKind, o.item.id, o.n)
    ask()
  }, [round.key, started]) // eslint-disable-line react-hooks/exhaustive-deps

  const nextRound = () => {
    stopAudio()
    setWrong(new Set())
    setSolved(false)
    setRound(makeRound())
  }

  const pick = async (o: Item) => {
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
    if (kind === 'animal' && SOUND_CREDITS[o.id]?.kind === 'real') seq.push(soundUrl(o.id))
    const finished = await playSequence(seq, undefined, 150)
    if (finished) {
      await new Promise((r) => setTimeout(r, 500))
      nextRound()
    }
  }

  const promptName = round.target.name.toLowerCase()

  return (
    <Screen tone={kind}>
      <TopBar kind={kind} mode="kuis" />
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
                {quizPrompt}{' '}
                <b>{promptName}</b>?
              </h2>
              <span className="kg-stars" aria-label={`${stars} bintang`}>
                ⭐ {stars}
              </span>
            </div>
            <div className={`kg-options kg-options-${optionLayout}`} key={round.key}>
              {round.options.map(({ item: o, n }) => {
                const isWrong = wrong.has(o.id)
                const isRight = solved && o.id === round.target.id
                return (
                  <button
                    key={o.id}
                    className={`kg-option${isWrong ? ' wrong' : ''}${isRight ? ' right' : ''}${solved && !isRight ? ' faded' : ''}`}
                    onClick={() => pick(o)}
                    aria-label="Pilihan"
                  >
                    {photoKind ? <Photo kind={photoKind} id={o.id} n={n} /> : <Swatch color={asColor(o)} />}
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

function CreditList({ kind, items }: { kind: PhotoKind; items: Item[] }) {
  return (
    <ul>
      {items.map((it) => {
        const photos = PHOTO_CREDITS[it.id] ?? []
        const sound = kind === 'animal' ? SOUND_CREDITS[it.id] : undefined
        return (
          <li key={it.id}>
            <img src={photoUrl(kind, it.id, 1)} alt="" loading="lazy" />
            <div>
              <b>{it.name}</b>
              {photos.map((c, i) => (
                <CreditLine key={i} label={photos.length > 1 ? `Foto ${i + 1}` : 'Foto'} c={c} />
              ))}
              <CreditLine label={sound?.kind === 'ambient' ? 'Suara (suasana)' : 'Suara'} c={sound} />
            </div>
          </li>
        )
      })}
    </ul>
  )
}

function Credits() {
  const sections = useMemo(
    () => [
      { title: 'Hewan', kind: 'animal' as const, items: ANIMALS },
      { title: 'Buah', kind: 'fruit' as const, items: FRUITS },
    ],
    [],
  )
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
          Foto hewan dan buah serta suara hewan berasal dari Wikimedia Commons dan dipakai sesuai lisensinya.
          Ilustrasi emosi dibuat untuk permainan ini. Suara pembaca nama dibuat dengan suara komputer berbahasa Indonesia.
        </p>
        {sections.map((s) => (
          <section key={s.kind}>
            <h2>{s.title}</h2>
            <CreditList kind={s.kind} items={s.items} />
          </section>
        ))}
      </main>
    </Screen>
  )
}
