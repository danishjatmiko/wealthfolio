import photoCredits from './photoCredits.json'
import soundCredits from './soundCredits.json'

// Everything the game needs lives as static files under public/, so the
// page works without the API and without a login.
export const BASE = '/interactive-game'

export interface Animal {
  id: string
  name: string
}

export interface GameColor {
  id: string
  name: string
  hex: string
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

export const ANIMALS: Animal[] = [
  { id: 'dog', name: 'Anjing' },
  { id: 'chicken', name: 'Ayam' },
  { id: 'cat', name: 'Kucing' },
  { id: 'duck', name: 'Bebek' },
  { id: 'rhino', name: 'Badak' },
  { id: 'tiger', name: 'Harimau' },
  { id: 'leopard', name: 'Macan' },
  { id: 'lion', name: 'Singa' },
  { id: 'giraffe', name: 'Jerapah' },
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

export const COLORS: GameColor[] = [
  { id: 'kuning', name: 'Kuning', hex: '#FFD21F' },
  { id: 'hijau', name: 'Hijau', hex: '#2DB84C' },
  { id: 'merah', name: 'Merah', hex: '#E5322D' },
  { id: 'biru', name: 'Biru', hex: '#1F6FE5' },
  { id: 'oranye', name: 'Oranye', hex: '#FF8A00' },
  { id: 'coklat', name: 'Coklat', hex: '#8B5A2B' },
  { id: 'hitam', name: 'Hitam', hex: '#151515' },
  { id: 'putih', name: 'Putih', hex: '#FFFFFF' },
  { id: 'abu', name: 'Abu-abu', hex: '#9A9A9A' },
  { id: 'ungu', name: 'Ungu', hex: '#8E3FC4' },
  { id: 'pink', name: 'Pink', hex: '#FF6FB5' },
]

const AQUATIC = new Set(['koi', 'goldfish', 'shark', 'ray', 'seahorse'])

export const photoUrl = (id: string) => `${BASE}/animals/${id}.jpg`
export const soundUrl = (id: string) => `${BASE}/sounds/${id}.m4a`
export const voiceUrl = (key: string) => `${BASE}/voice/${key}.m4a`
export const letterUrl = (ch: string) => `${BASE}/voice/letters/${ch.toLowerCase()}.m4a`

export const PHOTO_CREDITS = photoCredits as Record<string, Credit>
export const SOUND_CREDITS = soundCredits as Record<string, SoundCredit>

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
