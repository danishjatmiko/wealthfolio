import photoCredits from './photoCredits.json'
import soundCredits from './soundCredits.json'

// Everything the game needs lives as static files under public/, so the
// page works without the API and without a login.
export const BASE = '/interactive-game'

export type PhotoKind = 'animal' | 'fruit'

export interface Item {
  id: string
  name: string
}

export interface GameColor extends Item {
  /** Any CSS background — a flat colour, or a gradient for Emas/Perak. */
  fill: string
}

export interface Credit {
  title: string
  author: string
  license: string
  licenseUrl: string
  sourceUrl: string
}

export interface SoundCredit extends Credit {
  kind: 'real' | 'ambient'
}

export const ANIMALS: Item[] = [
  { id: 'dog', name: 'Anjing' },
  { id: 'chicken', name: 'Ayam' },
  { id: 'cat', name: 'Kucing' },
  { id: 'duck', name: 'Bebek' },
  { id: 'rhino', name: 'Badak' },
  { id: 'tiger', name: 'Harimau' },
  { id: 'leopard', name: 'Macan' },
  { id: 'lion', name: 'Singa' },
  { id: 'giraffe', name: 'Jerapah' },
  { id: 'monkey', name: 'Monyet' },
  { id: 'gorilla', name: 'Gorila' },
  { id: 'hippo', name: 'Kuda Nil' },
  { id: 'cow', name: 'Sapi' },
  { id: 'goat', name: 'Kambing' },
  { id: 'sheep', name: 'Domba' },
  { id: 'buffalo', name: 'Kerbau' },
  { id: 'koi', name: 'Ikan Koi' },
  { id: 'goldfish', name: 'Ikan Mas' },
  { id: 'whale', name: 'Ikan Paus' },
  { id: 'shark', name: 'Ikan Hiu' },
  { id: 'ray', name: 'Ikan Pari' },
  { id: 'cockatoo', name: 'Burung Kakaktua' },
  { id: 'eagle', name: 'Burung Elang' },
  { id: 'snail', name: 'Keong' },
  { id: 'worm', name: 'Cacing' },
  { id: 'snake', name: 'Ular' },
  { id: 'rabbit', name: 'Kelinci' },
  { id: 'gecko', name: 'Cicak' },
  { id: 'crocodile', name: 'Buaya' },
  { id: 'seahorse', name: 'Kuda Laut' },
  { id: 'seal', name: 'Anjing Laut' },
  { id: 'pig', name: 'Babi' },
  { id: 'mosquito', name: 'Nyamuk' },
  { id: 'fly', name: 'Lalat' },
  { id: 'ant', name: 'Semut' },
  { id: 'bee', name: 'Lebah' },
]

export const FRUITS: Item[] = [
  { id: 'apel', name: 'Apel' },
  { id: 'pisang', name: 'Pisang' },
  { id: 'jeruk', name: 'Jeruk' },
  { id: 'mangga', name: 'Mangga' },
  { id: 'semangka', name: 'Semangka' },
  { id: 'anggur', name: 'Anggur' },
  { id: 'nanas', name: 'Nanas' },
  { id: 'pepaya', name: 'Pepaya' },
  { id: 'stroberi', name: 'Stroberi' },
  { id: 'melon', name: 'Melon' },
  { id: 'durian', name: 'Durian' },
  { id: 'rambutan', name: 'Rambutan' },
  { id: 'manggis', name: 'Manggis' },
  { id: 'salak', name: 'Salak' },
  { id: 'alpukat', name: 'Alpukat' },
  { id: 'jambu', name: 'Jambu Biji' },
  { id: 'naga', name: 'Buah Naga' },
  { id: 'kelengkeng', name: 'Kelengkeng' },
  { id: 'belimbing', name: 'Belimbing' },
  { id: 'kiwi', name: 'Kiwi' },
]

export const COLORS: GameColor[] = [
  { id: 'kuning', name: 'Kuning', fill: '#FFD21F' },
  { id: 'hijau', name: 'Hijau', fill: '#2DB84C' },
  { id: 'merah', name: 'Merah', fill: '#E5322D' },
  { id: 'biru', name: 'Biru', fill: '#1F6FE5' },
  { id: 'oranye', name: 'Oranye', fill: '#FF8A00' },
  { id: 'coklat', name: 'Coklat', fill: '#8B5A2B' },
  { id: 'hitam', name: 'Hitam', fill: '#151515' },
  { id: 'putih', name: 'Putih', fill: '#FFFFFF' },
  { id: 'abu', name: 'Abu-abu', fill: '#9A9A9A' },
  { id: 'ungu', name: 'Ungu', fill: '#8E3FC4' },
  { id: 'pink', name: 'Pink', fill: '#FF6FB5' },
  { id: 'birumuda', name: 'Biru Muda', fill: '#7CC8F8' },
  { id: 'nila', name: 'Nila', fill: '#3B3BB0' },
  { id: 'toska', name: 'Toska', fill: '#14BFB0' },
  {
    id: 'emas',
    name: 'Emas',
    fill: 'linear-gradient(135deg, #A87A0C 0%, #F7D96B 30%, #C99A1E 50%, #FFF0A6 70%, #A87A0C 100%)',
  },
  {
    id: 'perak',
    name: 'Perak',
    fill: 'linear-gradient(135deg, #8A8F96 0%, #EEF1F4 30%, #A7ADB4 50%, #FFFFFF 70%, #8A8F96 100%)',
  },
]

const AQUATIC = new Set(['koi', 'goldfish', 'shark', 'ray', 'seahorse'])

/** Every photo's credit, keyed by item id; the array index is the photo
 *  number minus one, so its length is how many photos the item has. */
export const PHOTO_CREDITS = photoCredits as Record<string, Credit[]>
export const SOUND_CREDITS = soundCredits as Record<string, SoundCredit>

const PHOTO_DIR: Record<PhotoKind, string> = { animal: 'animals', fruit: 'fruits' }

export const photoCount = (id: string) => PHOTO_CREDITS[id]?.length ?? 1
export const photoUrl = (kind: PhotoKind, id: string, n: number) => `${BASE}/${PHOTO_DIR[kind]}/${id}-${n}.jpg`
export const soundUrl = (id: string) => `${BASE}/sounds/${id}.m4a`
export const voiceUrl = (key: string) => `${BASE}/voice/${key}.m4a`
export const letterUrl = (ch: string) => `${BASE}/voice/letters/${ch.toLowerCase()}.m4a`

/** A random photo number for an item, avoiding `not` when there's a choice. */
export function randomPhoto(id: string, not?: number): number {
  const count = photoCount(id)
  if (count <= 1) return 1
  let n = not
  while (n === not) n = 1 + Math.floor(Math.random() * count)
  return n!
}

/** For animals that don't really make a sound, the sound button plays an
 *  ambient clip instead — this says so, so nobody teaches a child that a
 *  goldfish goes "blub blub". Null when the clip is the animal itself. */
export function ambientNote(id: string): string | null {
  if (SOUND_CREDITS[id]?.kind !== 'ambient') return null
  return AQUATIC.has(id)
    ? 'Hewan ini tidak bersuara — ini suara air 🫧'
    : 'Hewan ini tidak bersuara — ini suara di sekitarnya 🍃'
}

export function shuffle<T>(arr: T[]): T[] {
  const a = arr.slice()
  for (let i = a.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    ;[a[i], a[j]] = [a[j], a[i]]
  }
  return a
}
