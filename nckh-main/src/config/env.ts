import { z } from 'zod'

const envSchema = z.object({
    VITE_API_BASE_URL: z.string().url(),
    VITE_APP_NAME: z.string().default('My App'),
    VITE_APP_VERSION: z.string().default('1.0.0'),
})

function validateEnv() {
    try {
        return envSchema.parse(import.meta.env)
    } catch (error) {
        console.error('❌ Invalid environment variables:', error)
        throw new Error('Invalid environment variables')
    }
}

export const env = validateEnv()
