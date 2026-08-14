import { BrandLoader } from '../components/BrandLoader';

export function PublicLoadingScreen() {
  return (
    <main className="public-loading" aria-labelledby="public-loading-title">
      <div className="public-loading__content">
        <BrandLoader />
        <h1 id="public-loading-title">Открываем публичные итоги</h1>
        <p>Проверяем безопасную публичную карточку</p>
        <span className="visually-hidden" role="status">
          Публичные итоги загружаются.
        </span>
      </div>
    </main>
  );
}

export function PublicFailureScreen({ message }: { message: string }) {
  return (
    <main className="public-loading public-loading--failed" aria-labelledby="public-failure-title">
      <div className="public-loading__content">
        <span className="public-loading__failure-mark" aria-hidden="true">
          !
        </span>
        <h1 id="public-failure-title">Публичные итоги не открылись</h1>
        <p>{message}</p>
        <a className="public-loading__action" href="/">
          Создать свои итоги
        </a>
      </div>
    </main>
  );
}
