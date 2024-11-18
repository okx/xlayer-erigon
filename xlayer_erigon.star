# Kurtosis Starlark script to deploy xlayer-erigon components

def run(plan):
    # Define common image and user settings
    image = "xlayer-erigon:latest"
    user_id = User("1000:1000")

    # 1. Erigon Service
    erigon_service = plan.add_service(
        name="erigon",
        config=ServiceConfig(
            image=image,
            entrypoint=[
                "erigon",
                "--private.api.addr=0.0.0.0:9090",
                "--sentry.api.addr=sentry:9091",
                "--downloader.api.addr=downloader:9093",
                "--txpool.disable",
                "--metrics",
                "--metrics.addr=0.0.0.0",
                "--metrics.port=6060",
                "--pprof",
                "--pprof.addr=0.0.0.0",
                "--pprof.port=6061",
                "--authrpc.jwtsecret=/home/erigon/.local/share/erigon/jwt.hex",
                "--datadir=/home/erigon/.local/share/erigon"
            ],
            ports={
                "8551": 8551,
                "6060": 6060,
                "6061": 6061
            },
            user=user_id
        )
    )

    # 2. Sentry Service
    sentry_service = plan.add_service(
        name="sentry",
        config=ServiceConfig(
            image=image,
            entrypoint=[
                "sentry",
                "--sentry.api.addr=0.0.0.0:9091",
                "--datadir=/home/erigon/.local/share/erigon"
            ],
            ports={
                "30303/tcp": 30303,
                "30303/udp": 30303
            },
            user=user_id
        )
    )

    # 3. Downloader Service
    downloader_service = plan.add_service(
        name="downloader",
        config=ServiceConfig(
            image=image,
            entrypoint=[
                "downloader",
                "--downloader.api.addr=0.0.0.0:9093",
                "--datadir=/home/erigon/.local/share/erigon"
            ],
            ports={
                "42069/tcp": 42069,
                "42069/udp": 42069
            },
            user=user_id
        )
    )

    # 4. TxPool Service
    txpool_service = plan.add_service(
        name="txpool",
        config=ServiceConfig(
            image=image,
            entrypoint=[
                "txpool",
                "--private.api.addr=erigon:9090",
                "--txpool.api.addr=0.0.0.0:9094",
                "--sentry.api.addr=sentry:9091",
                "--datadir=/home/erigon/.local/share/erigon"
            ],
            ports={
                "9094": 9094
            },
            user=user_id
        )
    )

    # 5. RPCDaemon Service
    rpcdaemon_service = plan.add_service(
        name="rpcdaemon",
        config=ServiceConfig(
            image=image,
            entrypoint=[
                "rpcdaemon",
                "--http.addr=0.0.0.0",
                "--http.vhosts=any",
                "--http.corsdomain=*",
                "--ws",
                "--private.api.addr=erigon:9090",
                "--txpool.api.addr=txpool:9094",
                "--datadir=/home/erigon/.local/share/erigon"
            ],
            ports={
                "8545": 8545
            },
            user=user_id
        )
    )

    # Wait until all services are up and healthy
    plan.wait([erigon_service, sentry_service, downloader_service, txpool_service, rpcdaemon_service])
    
    # Log messages to indicate successful start
    plan.log("All xlayer-erigon services have started successfully in Kurtosis environment.")
