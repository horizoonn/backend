import { useMemo, useState } from 'react';
import type { CSSProperties } from 'react';

import type { DayActivity } from '../api/types';
import { buildYearRows, WEEKS_IN_MONTH } from '../lib/recapLogic';
import type { YearMonthRow } from '../lib/recapLogic';

interface YearBoardProps {
  days: DayActivity[];
  peak?: DayActivity | null;
}

const MONTHS_SHORT = [
  'ЯНВ',
  'ФЕВ',
  'МАР',
  'АПР',
  'МАЙ',
  'ИЮН',
  'ИЮЛ',
  'АВГ',
  'СЕН',
  'ОКТ',
  'НОЯ',
  'ДЕК',
];

const MONTHS_FULL = [
  'Январь',
  'Февраль',
  'Март',
  'Апрель',
  'Май',
  'Июнь',
  'Июль',
  'Август',
  'Сентябрь',
  'Октябрь',
  'Ноябрь',
  'Декабрь',
];

export function YearBoard({ days, peak }: YearBoardProps) {
  const [selected, setSelected] = useState<number | null>(null);

  const rows = useMemo(() => buildYearRows(days, peak ?? null), [days, peak]);

  if (rows.length === 0) {
    return null;
  }

  const highlighted = selected;
  const highlightedRow = highlighted === null ? null : (rows[highlighted] ?? null);
  const caption = highlightedRow ? monthCaption(highlightedRow) : peakCaption(peak ?? null);

  return (
    <div className="year-board">
      <div className="year-board__scene">
        <div className="year-board__plate">
          <ul className="year-board__rows">
            {rows.map((row) => (
              <li key={row.month}>
                <button
                  type="button"
                  className={`year-board__row${highlighted === row.month ? ' year-board__row--active' : ''}${
                    highlighted !== null && highlighted !== row.month
                      ? ' year-board__row--muted'
                      : ''
                  }`}
                  aria-pressed={selected === row.month}
                  aria-label={monthSummary(row)}
                  onClick={() =>
                    setSelected((current) => (current === row.month ? null : row.month))
                  }
                >
                  <span className="year-board__label" aria-hidden="true">
                    {MONTHS_SHORT[row.month]}
                  </span>

                  <span
                    className="year-board__cells"
                    aria-hidden="true"
                    style={{ '--columns': WEEKS_IN_MONTH } as CSSProperties}
                  >
                    {row.columns.map((column) => (
                      <span
                        key={column.key}
                        className={`year-board__cell${column.peak ? ' year-board__cell--peak' : ''}`}
                        data-level={column.level}
                        style={{ '--level': column.level } as CSSProperties}
                      >
                        <span className="year-board__bar">
                          <span className="year-board__bar-top" />
                        </span>
                      </span>
                    ))}
                  </span>
                </button>
              </li>
            ))}
          </ul>
        </div>
      </div>

      <div className="year-board__legend">
        <span className="year-board__legend-title">Действий за неделю</span>

        <span className="year-board__legend-scale">
          <span className="year-board__legend-bound">реже</span>
          <span className="year-board__legend-swatches" aria-hidden="true">
            {[1, 2, 3, 4].map((level) => (
              <span key={level} className="year-board__legend-cell" data-level={level} />
            ))}
          </span>
          <span className="year-board__legend-bound">чаще</span>
        </span>

        <span className="year-board__legend-peak">
          <span
            className="year-board__legend-cell year-board__legend-cell--peak"
            data-level="4"
            aria-hidden="true"
          />
          неделя рекорда
        </span>
      </div>

      <p className="year-board__caption" role="status">
        {caption.title ? <span className="year-board__caption-title">{caption.title}</span> : null}
        <span className="year-board__caption-text">{caption.text}</span>
      </p>
    </div>
  );
}

interface Caption {
  title: string | null;
  text: string;
}

function monthCaption(row: YearMonthRow): Caption {
  return {
    title: MONTHS_FULL[row.month] ?? '',
    text: `${formatActions(row.actions)} · ${formatActiveDays(row.activeDays)}`,
  };
}

function monthSummary(row: YearMonthRow): string {
  const { title, text } = monthCaption(row);

  return `${title}: ${text}`;
}

function peakCaption(peak: DayActivity | null): Caption {
  if (!peak) {
    return {
      title: null,
      text: 'Каждый столбик — семидневный период месяца. Выберите месяц, чтобы рассмотреть.',
    };
  }

  return {
    title: formatDate(peak.date),
    text: `самый активный день года · ${formatActions(peak.actions)}`,
  };
}

function formatNumber(value: number): string {
  return new Intl.NumberFormat('ru-RU').format(value);
}

function formatActions(actions: number): string {
  const tail = actions % 100 >= 11 && actions % 100 <= 14 ? 'действий' : pluralTail(actions % 10);

  return `${formatNumber(actions)} ${tail}`;
}

function pluralTail(lastDigit: number): string {
  if (lastDigit === 1) {
    return 'действие';
  }

  if (lastDigit >= 2 && lastDigit <= 4) {
    return 'действия';
  }

  return 'действий';
}

function formatActiveDays(activeDays: number): string {
  const tail =
    activeDays % 100 >= 11 && activeDays % 100 <= 14
      ? 'активных дней'
      : activeDays % 10 === 1
        ? 'активный день'
        : activeDays % 10 >= 2 && activeDays % 10 <= 4
          ? 'активных дня'
          : 'активных дней';

  return `${activeDays} ${tail}`;
}

function formatDate(date: string): string {
  const parsed = new Date(`${date}T00:00:00Z`);
  if (Number.isNaN(parsed.getTime())) {
    return date;
  }

  return new Intl.DateTimeFormat('ru-RU', {
    day: 'numeric',
    month: 'long',
    timeZone: 'UTC',
  }).format(parsed);
}
