import fs from 'node:fs/promises'
import path from 'node:path'
import sharp from 'sharp'

const rootDir = process.cwd()
const sourceDir = path.join(rootDir, 'src', 'assets')
const outputDir = path.join(sourceDir, 'optimized')

const heroVariants = [
  { format: 'avif', options: { quality: 50, effort: 6 }, width: 1600 },
  { format: 'webp', options: { quality: 72, effort: 6 }, width: 1600 },
]

const cardVariants = [
  { format: 'avif', options: { quality: 45, effort: 6 }, width: 640 },
  { format: 'webp', options: { quality: 68, effort: 6 }, width: 640 },
]

const imageJobs = [
  {
    key: 'thu',
    heroSource: 'qingrenpo-wop-v1.png',
    cardSource: 'thu.png',
  },
  {
    key: 'pku',
    heroSource: 'pku.jpg',
    cardSource: 'pku.jpg',
  },
  {
    key: 'bnu',
    heroSource: 'bnu.png',
    cardSource: 'bnu.png',
  },
  {
    key: 'ruc',
    heroSource: 'ruc.png',
    cardSource: 'ruc.png',
  },
  {
    key: 'cuc',
    heroSource: 'cuc.jpg',
    cardSource: 'cuc.jpg',
  },
  {
    key: 'bfsu',
    heroSource: 'bfsu.jpg',
    cardSource: 'bfsu.jpg',
  },
]

async function ensureDir(dirPath) {
  await fs.mkdir(dirPath, { recursive: true })
}

async function renderVariant(inputPath, outputPath, variant) {
  const pipeline = sharp(inputPath).rotate().resize({
    width: variant.width,
    withoutEnlargement: true,
  })

  if (variant.format === 'avif') {
    await pipeline.avif(variant.options).toFile(outputPath)
    return
  }

  await pipeline.webp(variant.options).toFile(outputPath)
}

async function main() {
  const heroesDir = path.join(outputDir, 'heroes')
  const cardsDir = path.join(outputDir, 'cards')

  await Promise.all([ensureDir(heroesDir), ensureDir(cardsDir)])

  for (const job of imageJobs) {
    const heroInput = path.join(sourceDir, job.heroSource)
    const cardInput = path.join(sourceDir, job.cardSource)

    for (const variant of heroVariants) {
      const outputPath = path.join(heroesDir, `${job.key}.${variant.format}`)
      await renderVariant(heroInput, outputPath, variant)
    }

    for (const variant of cardVariants) {
      const outputPath = path.join(cardsDir, `${job.key}.${variant.format}`)
      await renderVariant(cardInput, outputPath, variant)
    }
  }

  console.log(`Optimized images written to ${path.relative(rootDir, outputDir)}`)
}

main().catch((error) => {
  console.error(error)
  process.exitCode = 1
})
