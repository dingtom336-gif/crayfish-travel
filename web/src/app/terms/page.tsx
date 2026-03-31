export default function TermsPage() {
  return (
    <main className="bg-gray-50 py-12 flex-1">
      <div className="max-w-3xl mx-auto px-4">
        <h1 className="text-2xl font-bold mb-8">服务条款</h1>
        <div className="bg-white rounded-xl p-8 shadow-sm space-y-6 text-sm text-gray-600 leading-relaxed">
          <section>
            <h2 className="text-lg font-semibold text-gray-800 mb-3">1. 服务说明</h2>
            <p>小龙虾旅行是一个旅行需求撮合平台，为用户匹配优质旅行方案。本平台不直接提供旅行服务，而是连接用户与合格的旅行服务供应商。</p>
          </section>
          <section>
            <h2 className="text-lg font-semibold text-gray-800 mb-3">2. 用户责任</h2>
            <p>用户应提供真实、准确的个人信息和旅行需求。用户理解并同意，平台基于用户提供的信息进行方案匹配，信息不准确可能影响方案质量。</p>
          </section>
          <section>
            <h2 className="text-lg font-semibold text-gray-800 mb-3">3. 费用说明</h2>
            <p>方案总价包含基础行程费用和退改权益服务费。退改权益服务费用于保障用户的退改权益，具体规则详见退款权益说明。</p>
          </section>
          <section>
            <h2 className="text-lg font-semibold text-gray-800 mb-3">4. 免责声明</h2>
            <p>本平台不是金融机构，不提供任何形式的保险产品。退改权益服务由平台风控资金池提供支持，具体退款规则和限制请参阅退款权益说明。</p>
          </section>
        </div>
      </div>
    </main>
  )
}
