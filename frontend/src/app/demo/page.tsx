import Image from 'next/image'
import Link from 'next/link'
import styles from './demo.module.css'

type Design = {
  name: string
  screenshot: string
  description: string
  path: string // route to navigate to when clicked (optional)
}

const designs: Design[] = [
  {
    name: 'Ethereal Monolith',
    screenshot: '/images/demo/ethereal-monolith.png',
    description: 'A immersive visualization module for strategy performance and market dynamics.',
    path: '/app/ethereal-monolith' // placeholder route
  },
  {
    name: 'Activity Feed',
    screenshot: '/images/demo/activity.png',
    description: 'Real-time activity log of agent actions, trades, and system events.',
    path: '/app/activity'
  },
  {
    name: 'Agent Settings',
    screenshot: '/images/demo/agent-settings.png',
    description: 'Configure risk parameters, leverage, position sizing, and execution logic.',
    path: '/app/agent-settings'
  },
  {
    name: 'Dashboard',
    screenshot: '/images/demo/dashboard.png',
    description: 'Main overview of portfolio performance, open positions, and key metrics.',
    path: '/app/dashboard'
  },
  {
    name: 'Portfolio',
    screenshot: '/images/demo/portfolio.png',
    description: 'Detailed view of holdings, allocations, profit/loss, and historical performance.',
    path: '/app/portfolio'
  },
]

export default function DemoPage() {
  return (
    <main className={styles.container}>
      <h1 className={styles.title}>O(Alpha) Product Preview</h1>
      <p className={styles.subtitle}>
        Explore the core components of the O(Alpha) trading platform. Click on any card to see a live
        version of the feature (where implemented) or view the design details.
      </p>
      <div className={styles.grid}>
        {designs.map((d) => (
          <Link key={d.name} href={d.path} passHref className={styles.card}>
            <Image
              src={d.screenshot}
              alt={d.name}
              width={400}
              height={250}
              className={styles.image}
            />
            <div className={styles.content}>
              <h2 className={styles.title2}>{d.name}</h2>
              <p className={styles.description}>{d.description}</p>
            </div>
          </Link>
        ))}
      </div>
      <div className={styles.footer}>
        <p>
          To experience the full interactive demo, please log in with a valid account or contact the
          team for access.
        </p>
      </div>
    </main>
  )
}