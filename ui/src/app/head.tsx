import Head from 'next/head'

function IndexPage() {
    return (
        <div>
            <Head>
                <title>My page title</title>
                <link rel="stylesheet"
                    href="https://unpkg.com/mvp.css" />ÔÔ
                <script src="https://accounts.google.com/gsi/client"
                    onLoad={() => {
                        console.log('TODO: add onload function')
                    }}>
                </script>
            </Head>
            <p>Hello world!</p>
        </div>
    )
}

export default IndexPage