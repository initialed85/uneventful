import argparse
import datetime
import json
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from threading import Event, RLock
from typing import Dict, Any, Optional

from requests import Session

_lock = RLock()
_count = 0


def _work(
        method: str,
        headers: Dict[str, Any],
        data: Optional[str],
        url: str,
        session: Session,
):
    global _lock, _count

    with session as s:
        _method = getattr(s, method)

        r = _method(
            url=url,
            headers=headers,
            data=json.dumps(data),
            timeout=5,
        )

        if r.status_code not in (200, 201):
            print(r.text)

        with _lock:
            _count += 1


def _loop(
        method: str,
        headers: Optional[str],
        data: Optional[str],
        url: str,
        period: float,
        stop_event: Event,
        just_once: bool = False,
):
    method = method.strip().lower()

    try:
        headers = json.loads(headers)
    except Exception:
        headers = None

    try:
        data = json.loads(data)
    except Exception:
        data = None

    period_delta = datetime.timedelta(seconds=period)

    while not stop_event.is_set():

        before = datetime.datetime.now()

        try:
            _work(
                method=method,
                headers=headers,
                data=data,
                url=url,
                session=Session(),
            )
        except Exception as e:
            print(repr(e))

        after = datetime.datetime.now()

        if just_once:
            break

        sleep_until = before + period_delta

        if after > sleep_until:
            continue

        sleep_seconds = (sleep_until - after).total_seconds()

        time.sleep(sleep_seconds)


def main():
    parser = argparse.ArgumentParser()

    parser.add_argument(
        "-s",
        dest="silent",
        required=False,
        default=False,
        action="store_true",
    )
    parser.add_argument(
        "-X",
        dest="method",
        type=str,
        required=False,
        default="GET",
    )
    parser.add_argument(
        "-H",
        dest="headers",
        type=str,
        required=False,
        default=None,
    )
    parser.add_argument(
        "-d",
        dest="data",
        type=str,
        required=False,
        default=None,
    )
    parser.add_argument(
        "url",
        type=str,
        nargs=1,
    )
    parser.add_argument(
        "--loop",
        dest="loop",
        required=False,
        default=False,
        action="store_true",
    )
    parser.add_argument(
        "--period",
        dest="period",
        type=float,
        required=False,
        default=0.1,
    )
    parser.add_argument(
        "--workers",
        dest="workers",
        type=int,
        required=False,
        default=1,
    )

    args = parser.parse_args()

    stop_event = Event()

    kwargs = dict(
        method=args.method,
        headers=args.headers,
        data=args.data,
        url=args.url[0],
        period=args.period,
        stop_event=stop_event,
    )

    executor = ThreadPoolExecutor(max_workers=args.workers)

    if not args.loop:
        _loop(
            just_once=True,
            **kwargs,
        )

        exit(0)

    futures = [
        executor.submit(
            _loop,
            **kwargs,
        )
        for _ in range(0, args.workers)
    ]

    last_count = 0

    while 1:
        with _lock:
            requests_per_second = _count - last_count
            last_count = _count

        print(f"{requests_per_second} requests per second")

        try:
            time.sleep(1)
        except KeyboardInterrupt:
            break

    stop_event.set()

    for future in as_completed(futures):
        _ = future.result()

    executor.shutdown()


if __name__ == "__main__":
    main()
