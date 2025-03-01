import { checkUrl } from "../scripts/url-check.js"

export const options = {
    scenarios: {
        availability: {
            executor: 'constant-vus',
            vus: 5,
            duration: "30s",

            tags: {
                app: 'test-app',
                env: 'development'
            }
        },
    },
    thresholds: {
        checks: ['rate==1.0'],
    },
}

export default function () {
    checkUrl();
}