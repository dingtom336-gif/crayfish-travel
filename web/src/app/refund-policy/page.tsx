export default function RefundPolicyPage() {
  return (
    <main className="bg-gray-50 py-12 flex-1">
      <div className="max-w-3xl mx-auto px-4">
        <h1 className="text-2xl font-bold mb-8">退款权益说明</h1>
        <div className="bg-white rounded-xl p-8 shadow-sm space-y-6 text-sm text-gray-600 leading-relaxed">
          <section>
            <h2 className="text-lg font-semibold text-gray-800 mb-3">1. 退改权益服务</h2>
            <p>每个旅行方案包含退改权益服务费，用于保障用户在特定条件下的退改权益。该服务由平台风控资金池提供支持，不属于保险产品。</p>
          </section>
          <section>
            <h2 className="text-lg font-semibold text-gray-800 mb-3">2. 退款条件</h2>
            <p>符合以下条件可申请退款：出行前申请、订单状态为已确认或已完成。退款金额由供应商退还比例和平台风控资金池补偿共同决定。</p>
          </section>
          <section>
            <h2 className="text-lg font-semibold text-gray-800 mb-3">3. 退款限制</h2>
            <p>为防止恶意退款，同一用户每月最多可发起3次退款申请。超出限制的退款申请将进入人工审核流程。</p>
          </section>
          <section>
            <h2 className="text-lg font-semibold text-gray-800 mb-3">4. 退款时效</h2>
            <p>退款申请提交后，将在3-5个工作日内处理完成。退款金额将原路返回至支付账户。</p>
          </section>
        </div>
      </div>
    </main>
  )
}
