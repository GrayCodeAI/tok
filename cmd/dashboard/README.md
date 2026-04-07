# TokMan Analytics Dashboard

Enterprise-grade analytics dashboard for token management and cost optimization.

## Features

- **Real-time Metrics**: Track token savings, compression ratios, and cost reduction
- **Team Analytics**: Per-team performance metrics and team member statistics
- **Trend Analysis**: Historical trends with daily/weekly/monthly granularity
- **Filter Performance**: Identify top-performing filters by effectiveness
- **Cost Projection**: Forecast monthly costs and savings
- **Responsive Design**: Mobile-friendly, works on all screen sizes
- **Dark Mode Ready**: Tailwind CSS with dark mode support

## Technology Stack

- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **UI**: React 18 with Tailwind CSS
- **Charts**: Recharts
- **State Management**: TanStack React Query + Zustand
- **HTTP Client**: Axios
- **Icons**: Lucide React

## Getting Started

### Prerequisites

- Node.js 18+
- npm or yarn

### Installation

```bash
cd cmd/dashboard
npm install
```

### Development

```bash
npm run dev
```

Dashboard will be available at `http://localhost:3000`

### Build

```bash
npm run build
npm start
```

### Type Checking

```bash
npm run type-check
```

## Environment Variables

Create `.env.local`:

```env
# API endpoint for analytics service
NEXT_PUBLIC_API_URL=http://localhost:8083

# Optional: Custom branding
NEXT_PUBLIC_APP_NAME=TokMan
```

## Project Structure

```
src/
├── app/                    # Next.js App Router
│   ├── layout.tsx         # Root layout
│   ├── page.tsx           # Dashboard home
│   ├── globals.css        # Global styles
│   └── api/               # API route handlers (optional)
├── components/
│   ├── dashboard/         # Dashboard sub-components
│   │   ├── metrics-cards.tsx
│   │   ├── trend-chart.tsx
│   │   ├── filter-effectiveness.tsx
│   │   ├── cost-projection.tsx
│   │   └── index.tsx
│   ├── header.tsx
│   ├── sidebar.tsx
│   ├── providers.tsx
│   └── loading-spinner.tsx
├── hooks/
│   └── use-dashboard.ts   # React Query hooks
├── lib/
│   └── api-client.ts      # HTTP client
├── types/
│   └── analytics.ts       # TypeScript types
└── store/
    └── team.ts            # Zustand stores
```

## API Integration

The dashboard connects to the TokMan Analytics gRPC service via HTTP/REST gateway.

### Required Endpoints

```
GET /dashboard?team_id=<team-id>
  Response:
  {
    teamStats: {...},
    economics: {...},
    trends: [...],
    topFilters: [...],
    projection: {...}
  }
```

## Customization

### Theming

Edit `tailwind.config.ts` to customize colors:

```ts
colors: {
  tokman: {
    50: '#f0f9ff',
    ...
  },
}
```

### Adding New Pages

1. Create new directory in `src/app/`
2. Add `page.tsx`
3. Update navigation in `src/components/sidebar.tsx`

## Performance

- Optimized with Next.js Image component
- React Query caching (5-minute stale time)
- Code splitting and lazy loading
- CSS-in-JS with Tailwind (no runtime overhead)

## Testing

```bash
npm test
npm run test:watch
```

## Deployment

### Vercel (Recommended)

```bash
vercel deploy
```

### Docker

```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY .next ./. next
EXPOSE 3000
CMD ["npm", "start"]
```

## Troubleshooting

### API Connection Issues

1. Verify Analytics service is running on port 8083
2. Check `NEXT_PUBLIC_API_URL` environment variable
3. Verify CORS headers on API gateway

### Data Not Loading

1. Check browser console for errors
2. Verify `team_id` is set correctly
3. Check Network tab in DevTools for API response

## Contributing

See main TokMan repository for contribution guidelines.

## License

MIT - See LICENSE in root directory
