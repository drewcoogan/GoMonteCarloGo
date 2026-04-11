import React, { useEffect, useState } from 'react';
import { ReactComponent as HeartSvg } from '../assets/heart.svg';
import { isServiceHealthy } from '../controllers/heartbeat';

const POLL_MS = 5000;

type HeartTone = 'healthy' | 'unhealthy' | 'pending';

const heartColor: Record<HeartTone, string> = {
  healthy: '#2e7d32',
  unhealthy: '#c62828',
  pending: '#bdbdbd',
};

function HeartIcon({ tone }: { tone: HeartTone }) {
  return (
    <HeartSvg
      width={22}
      height={22}
      aria-hidden
      style={{
        display: 'block',
        flexShrink: 0,
        color: heartColor[tone],
      }}
    />
  );
}

const ServiceHealthIndicator: React.FC = () => {
  const [healthy, setHealthy] = useState<boolean | null>(null);

  useEffect(() => {
    let cancelled = false;

    const tick = async () => {
      const ok = await isServiceHealthy();
      if (!cancelled) setHealthy(ok);
    };

    tick();
    const id = window.setInterval(tick, POLL_MS);
    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, []);

  const title =
    healthy === null
      ? 'Checking service…'
      : healthy
        ? 'API and database reachable'
        : 'Service unreachable or degraded';

  const tone: HeartTone = healthy === null ? 'pending' : healthy ? 'healthy' : 'unhealthy';

  return (
    <span
      title={title}
      aria-label={title}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <HeartIcon tone={tone} />
    </span>
  );
};

export default ServiceHealthIndicator;
