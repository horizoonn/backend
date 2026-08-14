import { ShieldCheck, Star } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';

import type { BadgeLevel } from '../api/types';
import styles from './BadgeCard.module.css';

const META: Record<BadgeLevel, { label: string; icon: LucideIcon }> = {
  bronze: { label: 'Бронза', icon: ShieldCheck },
  silver: { label: 'Серебро', icon: ShieldCheck },
  gold: { label: 'Золото', icon: Star },
};

interface BadgeCardProps {
  title: string;
  description?: string | null;
  level?: BadgeLevel | null;
  iconUrl?: string | null;
}

export function BadgeCard({ title, description, level, iconUrl }: BadgeCardProps) {
  const badgeLevel = level ?? 'bronze';
  const meta = META[badgeLevel];
  const Icon = meta.icon;

  return (
    <article className={styles.card}>
      <div className={styles.icon}>
        {iconUrl ? <img src={iconUrl} alt="" /> : <Icon aria-hidden="true" />}
      </div>
      <div className={styles.content}>
        <span className={styles.level}>{meta.label}</span>
        <h3 className={styles.title}>{title || 'Достижение'}</h3>
        {description ? <p className={styles.description}>{description}</p> : null}
      </div>
    </article>
  );
}
