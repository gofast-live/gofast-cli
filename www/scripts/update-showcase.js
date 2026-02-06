import fs from 'node:fs/promises';
import path from 'node:path';
import { createHighlighter } from 'shiki';

const GOFAST_APP_DIR = path.resolve('../../gofast-app');
const OUT_FILE = 'src/lib/data/showcase-code.json';

const FILES_TO_SHOWCASE = [
    {
        id: 'model_api',
        title: 'Generated API Handler',
        path: 'app/service-core/transport/skeleton/route.go',
        lang: 'go',
        description: 'Type-safe RPC handlers automatically generated for your data model.'
    },
    {
        id: 'model_service',
        title: 'Generated Service',
        path: 'app/service-core/domain/skeleton/service.go',
        lang: 'go',
        description: 'Clean business logic layer with validation, auth, and tracing hooks.'
    },
    {
        id: 'model_service_test',
        title: 'Generated Service Tests',
        path: 'app/service-core/domain/skeleton/service_test.go',
        lang: 'go',
        description: 'Comprehensive unit tests covering auth, validation, and data isolation.'
    },
    {
        id: 'model_ui',
        title: 'Generated UI Page',
        path: 'app/service-client/src/routes/(app)/models/skeletons/+page.svelte',
        lang: 'svelte',
        description: 'Responsive CRUD table with loading states and error handling.'
    },
    {
        id: 'model_e2e',
        title: 'Generated E2E Tests',
        path: 'e2e/skeletons.test.ts',
        lang: 'typescript',
        description: 'Full-stack Playwright tests generated for every model to ensure reliability.'
    },
    {
        id: 'otel',
        title: 'OpenTelemetry Setup',
        path: 'app/pkg/otel/otel.go',
        lang: 'go',
        description: 'Production-ready observability configuration for Traces, Metrics, and Logs.'
    }
];

async function main() {
    console.log('Generating showcase code...');
    
    const highlighter = await createHighlighter({
        themes: ['github-dark'],
        langs: ['go', 'sql', 'svelte', 'typescript'],
    });

    const output = [];

    for (const file of FILES_TO_SHOWCASE) {
        const fullPath = path.join(GOFAST_APP_DIR, file.path);
        
        try {
            let content = await fs.readFile(fullPath, 'utf-8');
            
            // Remove GF_ markers
            content = content.replace(/^\s*\/\/ GF_[A-Z_]+_(START|END)\n/gm, '');
            content = content.replace(/^\s*-- GF_[A-Z_]+_(START|END)\n/gm, '');
            content = content.replace(/^\s*<!-- GF_[A-Z_]+_(START|END) -->\n/gm, '');

            const html = highlighter.codeToHtml(content, {
                lang: file.lang,
                theme: 'github-dark'
            });

            output.push({
                ...file,
                html
            });
            
            console.log(`Processed ${file.path}`);
        } catch (err) {
            console.error(`Error processing ${file.path}:`, err);
        }
    }

    await fs.writeFile(OUT_FILE, JSON.stringify(output, null, 2));
    console.log(`Showcase data written to ${OUT_FILE}`);
}

main();